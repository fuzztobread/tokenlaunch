package queue

import (
	"context"
	"encoding/json"

	"github.com/IBM/sarama"

	"tokenlaunch/internal/domain"
)

type Kafka struct {
	producer sarama.SyncProducer
	topic    string
}

func NewKafka(brokers []string, topic string) (*Kafka, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &Kafka{
		producer: producer,
		topic:    topic,
	}, nil
}

func (k *Kafka) Publish(ctx context.Context, msg domain.Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	_, _, err = k.producer.SendMessage(&sarama.ProducerMessage{
		Topic: k.topic,
		Key:   sarama.StringEncoder(msg.ID),
		Value: sarama.ByteEncoder(data),
	})

	return err
}

func (k *Kafka) Close() error {
	return k.producer.Close()
}
