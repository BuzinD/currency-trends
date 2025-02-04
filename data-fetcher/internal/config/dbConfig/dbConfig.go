package dbConfig

import (
	"fmt"
	"log"

	"github.com/gofor-little/env"
)

const ENV_PATH = "env/db.env"

type DbConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DbName   string
}

func LoadEnv() {
	err := env.Load(ENV_PATH)
	if err != nil {
		log.Fatal(err)
	}
}

func GetDbConfig() (*DbConfig, error) {
	config := DbConfig{
		env.Get(Host, ""),
		env.Get(Port, ""),
		env.Get(User, ""),
		env.Get(Password, ""),
		env.Get(DbName, ""),
	}

	if config.Host == "" || config.Port == "" || config.User == "" || config.Password == "" || config.DbName == "" {
		return nil, fmt.Errorf("missing required environment variables host: %v port: %v user: %v passwordLength: %d dbName: %v",
			config.Host, config.Port, config.User, len(config.Password), config.DbName,
		)
	}

	return &config, nil
}

func GetTestDbConfig() (*DbConfig, error) {
	config := DbConfig{
		env.Get(TestHost, ""),
		env.Get(TestPort, ""),
		env.Get(TestUser, ""),
		env.Get(TestPassword, ""),
		env.Get(TestDbName, ""),
	}

	if config.Host == "" || config.Port == "" || config.User == "" || config.Password == "" || config.DbName == "" {
		return nil, fmt.Errorf("missing required environment variables host: %v port: %v user: %v passwordLength: %d dbName: %v",
			config.Host, config.Port, config.User, len(config.Password), config.DbName,
		)
	}

	return &config, nil
}
