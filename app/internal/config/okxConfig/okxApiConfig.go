package okxConfig

import (
	"fmt"
	"log"
	"strings"

	"github.com/gofor-little/env"
)

const API_ENV_PATH = "env/okx.env"

type OkxApiConfig struct {
	ApiKey         string
	Secret         string
	PassPhrase     string
	ApiUri         string
	CandlesPath    string
	TickersPath    string
	CurrenciesPath string
}

func LoadEnv() {
	err := env.Load(API_ENV_PATH)

	if err != nil {
		log.Fatal(err)
	}
}

func GetOkxApiConfig() (*OkxApiConfig, error) {

	config := OkxApiConfig{
		ApiKey:         env.Get(ApiKey, ""),
		Secret:         env.Get(Secret, ""),
		PassPhrase:     env.Get(PassPhrase, ""),
		ApiUri:         strings.Trim(env.Get(ApiUri, ""), "'\""),
		CandlesPath:    strings.Trim(env.Get(CandlesPath, ""), "'\""),
		TickersPath:    strings.Trim(env.Get(TickersPath, ""), "'\""),
		CurrenciesPath: strings.Trim(env.Get(CurrenciesPath, ""), "'\""),
	}

	if config.ApiKey == "" || config.Secret == "" || config.PassPhrase == "" {
		return nil, fmt.Errorf("missing required environment variables %v", config)
	}

	return &config, nil
}
