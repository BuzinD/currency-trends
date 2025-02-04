package okx

import (
	"crypto/hmac"
	"crypto/sha256"
	"cur/internal/config"
	"cur/internal/config/okxConfig"
	"cur/internal/store"
	structure "cur/internal/structure/response"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type OkxService struct {
	currencyRepository *store.CurrencyRepository
}

func NewOkxService(currencyRepository *store.CurrencyRepository) *OkxService {
	return &OkxService{currencyRepository: currencyRepository}
}

func (service *OkxService) UpdateCurrencies(okxConfig *okxConfig.OkxApiConfig) error {
	data, err := fetchCurrencies(okxConfig)
	if err != nil {
		return err
	}
	return service.currencyRepository.InsertOrUpdateCurrencies(data)
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

func UpdateTickers(config *config.Config) error {
	return fetchTickers(config)
}

func (service *OkxService) UpdateCandles() {
	url := "https://www.okx.com/api/v5/market/candles?instId=BTC-USDT&bar=1D"

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
}

func fetchTickers(config *config.Config) error {

	okxApiConfig := config.OkxApiConfig()
	req := okxApiConfig.ApiUri + okxApiConfig.TickersPath
	// req := "https://www.okx.com/api/v5/market/tickers?instType=SPOT"
	fmt.Println(req)

	resp, err := http.Get(okxApiConfig.ApiUri + okxApiConfig.TickersPath)

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
