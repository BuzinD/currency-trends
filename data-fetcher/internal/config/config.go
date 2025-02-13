package config

import (
	"cur/internal/config/dbConfig"
	"cur/internal/config/kafkaConfig"
	"cur/internal/config/okxConfig"
	"fmt"
)

type Config struct {
	okxConfig   *okxConfig.OkxApiConfig
	dbConfig    *dbConfig.DbConfig
	kafkaConfig *kafkaConfig.KafkaConfig
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

func (c *Config) DbConfig() *dbConfig.DbConfig {
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

func (c *Config) KafkaConfig() *kafkaConfig.KafkaConfig {
	if c.kafkaConfig == nil {
		var err error
		c.kafkaConfig, err = kafkaConfig.GetKafkaConfig()
		if err != nil {
			fmt.Errorf("Error getting kafkaConfig")
		}
	}

	return c.kafkaConfig
}

func LoadEnvs() {
	okxConfig.LoadEnv()
	dbConfig.LoadEnv()
	kafkaConfig.LoadEnv()
}
