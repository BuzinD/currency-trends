package config

import (
	"cur/internal/config/dbConfig"
	"cur/internal/config/okxConfig"
	"fmt"
)

type Config struct {
	okxConfig *okxConfig.OkxApiConfig
	dbConfig  *dbConfig.DbConfig
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) OkxApiConfig() *okxConfig.OkxApiConfig {
	if c.okxConfig != nil {
		return c.okxConfig
	}

	var err error
	c.okxConfig, err = okxConfig.GetOkxApiConfig()

	if err != nil {
		fmt.Errorf("Error getting okxConfig")
	}

	return c.okxConfig
}

func DbConfig(c *Config) *dbConfig.DbConfig {
	if c.dbConfig != nil {
		return c.dbConfig
	}
	var err error
	c.dbConfig, err = dbConfig.GetDbConfig()

	if err != nil {
		fmt.Errorf("Error getting okxConfig")
	}

	return c.dbConfig
}

func LoadEnvs() {
	okxConfig.LoadEnv()
	dbConfig.LoadEnv()
}
