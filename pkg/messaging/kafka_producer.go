package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaProducer struct {
	producer *kafka.Producer
	topic    string
}

// NewKafkaProducer creates a new KafkaProducer instance
func NewKafkaProducer(brokers, topic string) (*KafkaProducer, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": brokers,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	return &KafkaProducer{
		producer: p,
		topic:    topic,
	}, nil
}

func (p *KafkaProducer) Produce(ctx context.Context, eventType string, data interface{}) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	event := struct {
		Type string      `json:"type"`
		Data interface{} `json:"data"`
	}{
		Type: eventType,
		Data: data,
	}

	jsonData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event failed: %w", err)
	}

	err = p.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &p.topic, Partition: kafka.PartitionAny},
		Value:          jsonData,
	}, nil)
	if err != nil {
		return fmt.Errorf("produce message failed: %w", err)
	}

	return nil
}

func (p *KafkaProducer) Close() {
	p.producer.Flush(5000)
	p.producer.Close()
}
