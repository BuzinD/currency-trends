package okx

import (
	"cur/internal/config/okxConfig"
	"cur/internal/infrastructure"
	"cur/internal/infrastructure/dbConnection"
	"cur/internal/model"
	"cur/internal/store"

	structure "cur/internal/structure/response"
	"database/sql"
	"encoding/json"
	"fmt"
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

func TestMain(m *testing.M) {
	db, _ = dbConnection.GetTestDbConnection()
	infrastructure.MigrateTestDb(db)
	storage = makeTestStore()
	currencyRep = storage.Currency()
	okxService = NewOkxService(currencyRep)

	// Выполнение тестов
	exitVal := m.Run()

	// Очистка ресурсов, если требуется
	storage.CloseConnection()

	// Завершение тестов с корректным кодом выхода
	os.Exit(exitVal)
}

func makeTestStore() *store.Store {
	return store.NewStore(db)
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

		// Create a mock OkxApiConfig
		okxApiConfig = &okxConfig.OkxApiConfig{
			ApiUri:         mockServer.URL,
			CurrenciesPath: "/currencies",
		}

		err := okxService.UpdateCurrencies(okxApiConfig)

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

	okxApiConfig = &okxConfig.OkxApiConfig{
		ApiUri:         mockServer.URL,
		CurrenciesPath: "/currencies",
	}

	_, err := fetchCurrencies(okxApiConfig)
	assert.Error(t, err)
}
