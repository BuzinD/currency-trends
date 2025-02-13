package kafka

import (
	"cur/internal/config/kafkaConfig"
	"log"
	"time"

	"github.com/IBM/sarama"
)

type KafkaAsyncProducer struct {
	producer sarama.AsyncProducer
}

func NewKafkaAsyncProducer(kafkaConfig *kafkaConfig.KafkaConfig) (*KafkaAsyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal // Only wait for leader ack
	config.Producer.Retry.Max = 5                      // Retry up to 5 times
	config.Producer.Return.Successes = false           // No success confirmations
	config.Producer.Return.Errors = true               // Listen for errors
	config.Producer.Timeout = 10 * time.Second         // Message delivery timeout

	producer, err := sarama.NewAsyncProducer(kafkaConfig.Brokers, config)
	if err != nil {
		return nil, err
	}

	// Handle errors asynchronously
	go func() {
		for err := range producer.Errors() {
			log.Printf("Failed to send message to Kafka: %v", err)
		}
	}()

	return &KafkaAsyncProducer{producer: producer}, nil
}

func (kp *KafkaAsyncProducer) SendMessage(topic, message string) {
	kp.producer.Input() <- &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	}
}

func (kp *KafkaAsyncProducer) Close() error {
	return kp.producer.Close()
}
