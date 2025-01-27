package okxConfig

import (
	"fmt"
	"log"

	"github.com/gofor-little/env"
)

const API_ENV_PATH = "env/okx.env"

type OkxApiConfig struct {
	ApiKey     string
	Secret     string
	PassPhrase string
}

func LoadEnv() {
	err := env.Load(API_ENV_PATH)

	if err != nil {
		log.Fatal(err)
	}
}

func GetOkxApiConfig() (*OkxApiConfig, error) {

	config := OkxApiConfig{
		env.Get(ApiKey, ""),
		env.Get(Secret, ""),
		env.Get(PassPhrase, ""),
	}

	if config.ApiKey == "" || config.Secret == "" || config.PassPhrase == "" {
		return nil, fmt.Errorf("missing required environment variables %v", config)
	}

	return &config, nil
}
