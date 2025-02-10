package okx

import (
	"crypto/hmac"
	"crypto/sha256"
	"cur/internal/config/okxConfig"
	"cur/internal/helper/price"
	"cur/internal/model"
	"cur/internal/store"
	structure "cur/internal/structure/response"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
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
}

func NewOkxService(
	currencyRepository *store.CurrencyRepository,
	candleRepository *store.CandleRepository,
	config *okxConfig.OkxApiConfig,
) *OkxService {
	return &OkxService{
		currencyRepository: currencyRepository,
		candleRepository:   candleRepository,
		okxConfig:          config,
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

func fetchCurrencies(okxConfig *okxConfig.OkxApiConfig) (*[]structure.CurrencyResponseData, error) {

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

	var currencyResponse structure.CurrencyResponse

	err = json.NewDecoder(resp.Body).Decode(&currencyResponse)

	if err != nil {
		return nil, fmt.Errorf("Failed to decode JSON: %v.", err)
	}

	return &currencyResponse.Data, nil
}

func (okx *OkxService) UpdateTickers() error {
	return fetchTickers(okx.okxConfig)
}

func (okx *OkxService) UpdateCandles() {
	for _, cur2 := range okx.okxConfig.Currencies {
		pair := cur2 + "-" + okx.okxConfig.BaseCurrency

		for {
			before := okx.getLastTsForPair(pair)
			fmt.Println(before)
			candles, err := okx.fetchCandles(pair, before, "")

			if err != nil {
				fmt.Println(err)
				break
			}
			if len(candles) == 0 {
				fmt.Println("fetched 0 candles. exit")
				break
			}

			err = okx.candleRepository.InsertCandles(&candles)
		}
	}
}

func (okx *OkxService) UpdateHistoricalCandles() {
	for _, cur2 := range okx.okxConfig.Currencies {
		pair := cur2 + "-" + okx.okxConfig.BaseCurrency
		minAfter := okx.getLastTsForPair(pair)
		after := strconv.FormatInt(time.Now().UnixMilli(), 10)
		for {
			fmt.Printf("fetching chunk candles for pair %s, earlier than %s\n", pair, after)

			candles, err := okx.fetchCandles(pair, "", after)

			if err != nil {
				fmt.Println(err)
				break
			}
			if len(candles) < Limit {
				fmt.Println("fetched 0 candles. exit")
				break
			}

			err = okx.candleRepository.InsertCandles(&candles)
			if err != nil {
				fmt.Println(err)
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
		fmt.Println(err)
		return BeforeCandles
	}
	return lastTimestamp
}

// getFirstTsForPair getting min timestamp for pair
func (okx *OkxService) getFirstTsForPair(pair string) string {
	lastTimestamp, err := okx.candleRepository.GetFirstTsForPair(pair)
	if err != nil {
		fmt.Println(err)
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

	var response structure.CandlesResponse

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

func fetchTickers(config *okxConfig.OkxApiConfig) error {

	req := config.ApiUri + config.TickersPath
	// req := "https://www.okx.com/api/v5/market/tickers?instType=SPOT"
	fmt.Println(req)

	resp, err := http.Get(config.ApiUri + config.TickersPath)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Bad response: %v", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return fmt.Errorf("Failed to read response body: %v", err)
	}

	var tickerResponse structure.TickerResponse

	err = json.Unmarshal(body, &tickerResponse)
	if err != nil {
		return fmt.Errorf("Failed to read response body: %v", err)
	}

	for _, ticker := range tickerResponse.Data {
		fmt.Printf("Ticker: %s, Last Price: %s, 24h High: %s, 24h Low: %s\n",
			ticker.InstId, ticker.Last, ticker.High24H, ticker.Low24H)
	}

	return nil
}

func createSignature(timestamp, method, path, body string, conf *okxConfig.OkxApiConfig) string {
	signaturePayload := timestamp + method + path + body
	mac := hmac.New(sha256.New, []byte(conf.Secret))
	mac.Write([]byte(signaturePayload))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

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
