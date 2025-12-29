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

type KafkaConsumer struct {
	group   sarama.ConsumerGroup
	topic   string
	handler func(msg domain.Message) error
}

func NewKafkaConsumer(brokers []string, groupID, topic string) (*KafkaConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	group, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	return &KafkaConsumer{
		group: group,
		topic: topic,
	}, nil
}

func (c *KafkaConsumer) Consume(ctx context.Context, handler func(msg domain.Message) error) error {
	c.handler = handler

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if err := c.group.Consume(ctx, []string{c.topic}, c); err != nil {
				return err
			}
		}
	}
}

func (c *KafkaConsumer) Close() error {
	return c.group.Close()
}

func (c *KafkaConsumer) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (c *KafkaConsumer) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (c *KafkaConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var message domain.Message
		if err := json.Unmarshal(msg.Value, &message); err != nil {
			continue
		}

		if err := c.handler(message); err != nil {
			continue
		}

		session.MarkMessage(msg, "")
	}
	return nil
}
