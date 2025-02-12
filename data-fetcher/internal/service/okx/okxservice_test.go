package okx

import (
	"cur/internal/config/dbConfig"
	"cur/internal/config/okxConfig"
	"cur/internal/infrastructure/dbConnection"
	"cur/internal/model"
	"cur/internal/store"
	structure "cur/internal/structure/response"
	"database/sql"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var (
	db           *sql.DB
	storage      *store.Store
	currencyRep  *store.CurrencyRepository
	okxService   *OkxService
	okxApiConfig *okxConfig.OkxApiConfig
)

func makeTestStore() *store.Store {
	dbConfig.LoadEnv()
	conf, err := dbConfig.GetDbConfig()

	if err != nil {
		panic(err)
	}
	db, _ = dbConnection.GetDbConnection(conf)
	return store.NewStore(db)
}

func TestMain(m *testing.M) {
	storage = makeTestStore()
	currencyRep = storage.Currency()

	okxApiConfig = &okxConfig.OkxApiConfig{
		ApiUri:         "http://localhost",
		CurrenciesPath: "/currencies",
	}

	okxService = NewOkxService(
		currencyRep,
		storage.Candle(),
		okxApiConfig,
		log.New(),
	)

	// Выполнение тестов
	exitVal := m.Run()

	// Очистка ресурсов, если требуется
	storage.CloseConnection()

	// Завершение тестов с корректным кодом выхода
	os.Exit(exitVal)
}

func TestOkxService_UpdateCurrencies(t *testing.T) {
	type expectation struct {
		currencies []model.Currency
	}

	testCases := []struct {
		name     string
		expected expectation
		response structure.CurrencyResponse
		params   struct {
			truncateAfterTest bool
		}
	}{
		{
			name: "Ok, 3 currencies",
			expected: expectation{
				[]model.Currency{
					{Code: "BTC", Chain: "ETH20", CanWithdraw: true, CanDeposit: true},
					{Code: "ETH", Chain: "ETH20", CanWithdraw: true, CanDeposit: true},
					{Code: "USDT", Chain: "ETH20", CanWithdraw: true, CanDeposit: true},
				},
			},
			response: structure.CurrencyResponse{
				Data: []structure.CurrencyResponseData{
					{
						Ccy:    "BTC",
						Chain:  "ETH20",
						CanDep: true,
						CanWd:  true,
					},
					{
						Ccy:    "ETH",
						Chain:  "ETH20",
						CanDep: true,
						CanWd:  true,
					},
					{
						Ccy:    "USDT",
						Chain:  "ETH20",
						CanDep: true,
						CanWd:  true,
					},
				},
			},
			params: struct{ truncateAfterTest bool }{truncateAfterTest: true},
		},
		{
			name: "Ok, 3 currencies but 2 the same with different in CanWithdraw ",
			expected: expectation{
				[]model.Currency{
					{Code: "BTC", Chain: "ETH20", CanWithdraw: false, CanDeposit: true},
					{Code: "USDT", Chain: "ETH20", CanWithdraw: true, CanDeposit: true},
				},
			},
			response: structure.CurrencyResponse{
				Data: []structure.CurrencyResponseData{
					{
						Ccy:    "BTC",
						Chain:  "ETH20",
						CanDep: true,
						CanWd:  true,
					},
					{
						Ccy:    "BTC",
						Chain:  "ETH20",
						CanDep: true,
						CanWd:  false,
					},
					{
						Ccy:    "USDT",
						Chain:  "ETH20",
						CanDep: true,
						CanWd:  true,
					},
				},
			},
			params: struct{ truncateAfterTest bool }{truncateAfterTest: true},
		},
		{
			name: "Ok, 2 currencies needs for nex test",
			expected: expectation{
				[]model.Currency{
					{Code: "BTC", Chain: "ETH20", CanWithdraw: true, CanDeposit: true},
					{Code: "USDT", Chain: "ETH20", CanWithdraw: true, CanDeposit: true},
				},
			},
			response: structure.CurrencyResponse{
				Data: []structure.CurrencyResponseData{
					{
						Ccy:    "BTC",
						Chain:  "ETH20",
						CanDep: true,
						CanWd:  true,
					},
					{
						Ccy:    "USDT",
						Chain:  "ETH20",
						CanDep: true,
						CanWd:  true,
					},
				},
			},
			params: struct{ truncateAfterTest bool }{truncateAfterTest: false},
		},
		{
			name: "Ok, 2 currencies should rewrite db data from previous test",
			expected: expectation{
				[]model.Currency{
					{Code: "BTC", Chain: "ETH20", CanWithdraw: false, CanDeposit: false},
					{Code: "USDT", Chain: "ETH20", CanWithdraw: false, CanDeposit: false},
				},
			},
			response: structure.CurrencyResponse{
				Data: []structure.CurrencyResponseData{
					{
						Ccy:    "BTC",
						Chain:  "ETH20",
						CanDep: false,
						CanWd:  false,
					},
					{
						Ccy:    "USDT",
						Chain:  "ETH20",
						CanDep: false,
						CanWd:  false,
					},
				},
			},
			params: struct{ truncateAfterTest bool }{truncateAfterTest: false},
		},
	}

	for _, testCase := range testCases {
		fmt.Printf("Test case: %s \n", testCase.name)
		// Create a mock server
		mockServer := startMockServer(testCase.response)
		defer mockServer.Close()

		okxService.SetConfig(&okxConfig.OkxApiConfig{
			ApiUri:         mockServer.URL,
			CurrenciesPath: "/currencies",
		})

		err := okxService.UpdateCurrencies()

		currencies, err := currencyRep.FetchAll()

		assert.Equal(t, len(testCase.expected.currencies), len(currencies))

		assert.NoError(t, err)

		for i, v := range currencies {
			assert.Equal(t, testCase.expected.currencies[i].Code, v.Code)
			assert.Equal(t, testCase.expected.currencies[i].CanDeposit, v.CanDeposit)
			assert.Equal(t, testCase.expected.currencies[i].CanWithdraw, v.CanWithdraw)
		}

		if testCase.params.truncateAfterTest {
			err := storage.TruncateTables([]string{"currencies"})
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func startMockServer(response structure.CurrencyResponse) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
}

func TestOkxService_FetchCurrenciesFails(t *testing.T) {
	wrongJsonBody := "wrong json body"
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(wrongJsonBody)
	}))

	defer mockServer.Close()

	_, err := fetchCurrencies(okxApiConfig)
	assert.Error(t, err)
}

func TestOkxService_UpdateCandles(t *testing.T) {
	type expectation struct {
		BdQty int
	}

	type Response struct {
		Code string     `json:"code"`
		Msg  string     `json:"msg"`
		Data [][]string `json:"data"`
	}

	testCases := []struct {
		name      string
		expected  expectation
		responses []Response
		params    struct {
			truncateAfterTest bool
		}
	}{
		{
			name: "Ok 10 values",
			expected: expectation{
				6 * 2, //3 responses return 2 values * 2 currencies
			},
			responses: []Response{
				{
					Data: [][]string{
						{"1738857600000", "97338.7", "97893.3", "95680", "96888.7", "3706.88063172", "358669057.793862051", "358669057.793862051", "0"},
						{"1738771200000", "98114.6", "99132.2", "96156", "97338.6", "6877.12302248", "672151013.621598455", "672151013.621598455", "1"},
					},
				},
				{
					Data: [][]string{
						{"1738684800000", "99357.6", "100780.9", "96136.2", "98114.6", "9829.71499174", "965894235.673523517", "965894235.673523517", "1"},
						{"1738598400000", "98832", "102500", "97843.9", "99355.3", "13214.68600385", "1320230602.460214531", "1320230602.460214531", "1"},
					},
				},
				{
					Data: [][]string{
						{"1738512000000", "99348.2", "99490", "91182.6", "98831.9", "28649.09520794", "2725336187.812171192", "2725336187.812171192", "1"},
						{"1738425600000", "102074.2", "102289", "98170", "99348.2", "6364.25862137", "636421873.029058697", "636421873.029058697", "1"},
					},
				},
				{
					Data: [][]string{},
				},
				{
					Data: [][]string{
						{"1738857600000", "97338.7", "97893.3", "95680", "96888.7", "3706.88063172", "358669057.793862051", "358669057.793862051", "0"},
						{"1738771200000", "98114.6", "99132.2", "96156", "97338.6", "6877.12302248", "672151013.621598455", "672151013.621598455", "1"},
					},
				},
				{
					Data: [][]string{
						{"1738684800000", "99357.6", "100780.9", "96136.2", "98114.6", "9829.71499174", "965894235.673523517", "965894235.673523517", "1"},
						{"1738598400000", "98832", "102500", "97843.9", "99355.3", "13214.68600385", "1320230602.460214531", "1320230602.460214531", "1"},
					},
				},
				{
					Data: [][]string{
						{"1738512000000", "99348.2", "99490", "91182.6", "98831.9", "28649.09520794", "2725336187.812171192", "2725336187.812171192", "1"},
						{"1738425600000", "102074.2", "102289", "98170", "99348.2", "6364.25862137", "636421873.029058697", "636421873.029058697", "1"},
					},
				},
				{
					Data: [][]string{},
				},
			},

			params: struct{ truncateAfterTest bool }{truncateAfterTest: true},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			var requestCount int = 0

			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				response := testCase.responses[requestCount]
				requestCount++
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(response)
			}))
			defer mockServer.Close()

			okxService.SetConfig(&okxConfig.OkxApiConfig{
				ApiUri:         mockServer.URL,
				CurrenciesPath: "/candles",
				BaseCurrency:   "USDT",
				Currencies:     []string{"BTC", "ETH"},
			})

			okxService.UpdateCandles()

			candles, err := storage.Candle().FetchAll()

			assert.NoError(t, err)
			assert.Equal(t, testCase.expected.BdQty, len(candles))

			if testCase.params.truncateAfterTest {
				err := storage.TruncateTables([]string{"candles"})
				if err != nil {
					fmt.Println(err)
				}
			}
		})
	}
}
