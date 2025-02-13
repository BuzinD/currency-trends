package kafkaConfig

import (
	"log"
	"strings"

	"github.com/gofor-little/env"
)

const ENV_PATH = "env/kafka.env"

type KafkaConfig struct {
	Brokers []string
}

func LoadEnv() {
	err := env.Load(ENV_PATH)
	if err != nil {
		log.Fatal(err)
	}
}

func GetKafkaConfig() (*KafkaConfig, error) {

	brokersString := strings.Trim(env.Get(Brokers, ""), "\n'")

	return &KafkaConfig{
		parseBrokers(brokersString),
	}, nil
}

func parseBrokers(brokersString string) []string {
	b := strings.Split(brokersString, ",")

	if len(b) == 0 {
		log.Fatalf("Invalid brokers: %s", brokersString)
	}

	return b
}
