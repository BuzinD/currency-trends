package okx

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"cur/internal/config/kafkaConfig"
	"cur/internal/config/okxConfig"
	"cur/internal/helper/price"
	"cur/internal/infrastructure/kafka"
	"cur/internal/model"
	"cur/internal/service/okx/request"
	"cur/internal/service/okx/response"
	"cur/internal/store"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"

	"net/http"
	"strconv"
	"time"
)

const (
	BeforeCandles string = "1577836800000"
	Limit                = 100
)

type OkxService struct {
	currencyRepository *store.CurrencyRepository
	candleRepository   *store.CandleRepository
	okxConfig          *okxConfig.OkxApiConfig
	kafkaConfig        *kafkaConfig.KafkaConfig
	log                *log.Logger
}

func NewOkxService(
	currencyRepository *store.CurrencyRepository,
	candleRepository *store.CandleRepository,
	config *okxConfig.OkxApiConfig,
	kafkaConf *kafkaConfig.KafkaConfig,
	log *log.Logger,
) *OkxService {
	return &OkxService{
		currencyRepository: currencyRepository,
		candleRepository:   candleRepository,
		okxConfig:          config,
		kafkaConfig:        kafkaConf,
		log:                log,
	}
}

func (okx *OkxService) SetConfig(okxConfig *okxConfig.OkxApiConfig) {
	okx.okxConfig = okxConfig
}

func (okx *OkxService) UpdateCurrencies() error {
	data, err := fetchCurrencies(okx.okxConfig)
	if err != nil {
		return err
	}
	return okx.currencyRepository.InsertOrUpdateCurrencies(data)
}

func fetchCurrencies(okxConfig *okxConfig.OkxApiConfig) (*[]response.CurrencyResponseData, error) {

	client := &http.Client{}

	req, err := http.NewRequest("GET", okxConfig.ApiUri+okxConfig.CurrenciesPath, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to create request: %v", err)
	}

	req = getAuthHeaders(req, okxConfig, okxConfig.CurrenciesPath)

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Bad response: %v.", resp.Status)
	}

	var currencyResponse response.CurrencyResponse

	err = json.NewDecoder(resp.Body).Decode(&currencyResponse)

	if err != nil {
		return nil, fmt.Errorf("Failed to decode JSON: %v.", err)
	}

	return &currencyResponse.Data, nil
}

func (okx *OkxService) UpdateCandles() {
	for _, cur2 := range okx.okxConfig.Currencies {
		pair := cur2 + "-" + okx.okxConfig.BaseCurrency

		for {
			before := okx.getLastTsForPair(pair)
			candles, err := okx.fetchCandles(pair, before, "")

			if err != nil {
				log.Error(err)
				break
			}
			if len(candles) == 0 {
				break
			}

			err = okx.candleRepository.InsertCandles(&candles)
			if err != nil {
				log.Error(err)
			}
		}
	}
}

func (okx *OkxService) UpdateHistoricalCandles() {
	for _, cur2 := range okx.okxConfig.Currencies {
		pair := cur2 + "-" + okx.okxConfig.BaseCurrency
		minAfter := okx.getLastTsForPair(pair)
		after := strconv.FormatInt(time.Now().UnixMilli(), 10)
		for {
			log.Infof("fetching chunk candles for pair %s, earlier than %s\n", pair, after)
			candles, err := okx.fetchCandles(pair, "", after)

			if err != nil {
				log.Error(err)
				break
			}
			if len(candles) < Limit {
				break
			}

			err = okx.candleRepository.InsertCandles(&candles)
			if err != nil {
				log.Error(err)
				break
			}

			after = okx.getFirstTsForPair(pair)
			if after <= minAfter {
				break
			}
		}
	}
}

// getLastTsForPair getting max timestamp for pair
func (okx *OkxService) getLastTsForPair(pair string) string {
	lastTimestamp, err := okx.candleRepository.GetLastTsForPair(pair)
	if err != nil {
		return BeforeCandles
	}
	return lastTimestamp
}

// getFirstTsForPair getting min timestamp for pair
func (okx *OkxService) getFirstTsForPair(pair string) string {
	lastTimestamp, err := okx.candleRepository.GetFirstTsForPair(pair)
	if err != nil {
		return strconv.FormatInt(time.Now().UnixMilli(), 10)
	}
	return lastTimestamp
}

func (okx *OkxService) fetchCandles(pair, before, after string) ([]model.Candle, error) {
	url := fmt.Sprintf("%s?instId=%s&bar=%s&limit=%s", okx.okxConfig.ApiUri+
		okx.okxConfig.CandlesPath, pair, okx.okxConfig.CandlesBar, strconv.Itoa(Limit))

	if before != "" {
		url += "&before=" + before
	}

	if after != "" {
		url += "&after=" + after
	}

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	var response response.CandlesResponse

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var candles []model.Candle
	var timestamp time.Time

	for _, c := range response.Data {
		timestampInt, _ := strconv.ParseInt(c[0], 10, 64)
		timestamp = time.UnixMilli(timestampInt).In(time.UTC)

		open, _ := price.ParsePrice(c[1])
		high, _ := price.ParsePrice(c[2])
		low, _ := price.ParsePrice(c[3])
		closePrice, _ := price.ParsePrice(c[4])
		volume, _ := price.ParsePrice(c[5])

		candles = append(candles, model.Candle{
			Pair:      pair,
			Timestamp: timestamp,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     closePrice,
			Volume:    volume,
		})
	}

	return candles, nil
}

// FetchTrades TODO needs to do something with results
func (okx *OkxService) FetchTrades(ctx context.Context) {

	var reconnectInterval time.Duration = 1 * time.Second
	kafkaProducer, err := kafka.NewKafkaAsyncProducer(okx.kafkaConfig)

	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer kafkaProducer.Close()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping FetchTrades...")
			return
		default:

			conn, _, err := websocket.DefaultDialer.Dial(okx.okxConfig.WssEndpoint, nil)
			if err != nil {
				log.Printf("Failed to connect to WebSocket: %v", err)
				time.Sleep(reconnectInterval)
				reconnectInterval *= 2
				if reconnectInterval > 60*time.Second {
					reconnectInterval = 60 * time.Second
				}
				continue
			}
			log.Println("Connected to WebSocket")
			reconnectInterval = 1 * time.Second

			// Defer connection close
			defer conn.Close()

			err = okx.subscribeToTrades(conn)
			if err != nil {
				log.Printf("Failed to subscribe: %v", err)
				continue
			}

			// Listen for messages
			done := make(chan struct{})
			go func() {
				defer close(done)
				if err := listenForTrades(ctx, conn, kafkaProducer); err != nil {
					return
				}
			}()

			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					log.Println("Stopping WebSocket connection...")
					_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
					return
				case <-ticker.C:
					// Send ping to keep connection alive
					err := conn.WriteMessage(websocket.PingMessage, []byte{})
					if err != nil {
						log.Printf("Ping error: %v", err)
						return
					}
				case <-done:
					log.Println("Connection closed, attempting to reconnect...")
					break
				}
			}
		}
	}
}

// Subscribe to trades
func (okx *OkxService) subscribeToTrades(conn *websocket.Conn) error {
	var requestArgs []request.Arg

	for _, v := range okx.okxConfig.Currencies {
		requestArg := request.Arg{Channel: "trades", InstId: fmt.Sprintf("%s-%s", v, okx.okxConfig.BaseCurrency)}
		requestArgs = append(requestArgs, requestArg)
	}

	subscription := request.SubscriptionMessage{
		Op:   "subscribe",
		Args: requestArgs,
	}

	msg, _ := json.Marshal(subscription)
	err := conn.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		return fmt.Errorf("subscription failed: %v", err)
	}
	log.Info("Subscribed to trades.")
	return nil
}

// listenForTrades Listen for trades in real time
func listenForTrades(ctx context.Context, conn *websocket.Conn, kafkaProducer *kafka.KafkaAsyncProducer) error {
	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping trade listener...")
			return nil
		default:
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("Error reading message: %v", err)
				return err
			}

			// Parse Trade Message
			var trade response.TradeMessage
			err = json.Unmarshal(message, &trade)
			if err != nil {
				log.Printf("JSON unmarshal error: %v", err)
				continue
			}

			kafkaProducer.SendMessage("trades", string(message))

			for _, data := range trade.Data {
				fmt.Printf("[%s] Trade ID: %s | Price: %s | Size: %s | Side: %s | Time: %s\n",
					trade.Arg.InstId,
					data.TradeID,
					data.Price,
					data.Size,
					data.Side,
					data.Time,
				)
			}
		}
	}
}

// createSignature create signature for okx request
func createSignature(timestamp, method, path, body string, conf *okxConfig.OkxApiConfig) string {
	signaturePayload := timestamp + method + path + body
	mac := hmac.New(sha256.New, []byte(conf.Secret))
	mac.Write([]byte(signaturePayload))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// getAuthHeaders make headers for okx api request
func getAuthHeaders(
	req *http.Request,
	conf *okxConfig.OkxApiConfig,
	path string,
) *http.Request {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	signature := createSignature(timestamp, "GET", path, "", conf)

	req.Header.Set("OK-ACCESS-KEY", conf.ApiKey)
	req.Header.Set("OK-ACCESS-SIGN", signature)
	req.Header.Set("OK-ACCESS-TIMESTAMP", timestamp)
	req.Header.Set("OK-ACCESS-PASSPHRASE", conf.PassPhrase)
	req.Header.Set("Content-Type", "application/json")

	return req
}
