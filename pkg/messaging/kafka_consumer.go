package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/metacode-dream-team/MetaCode/pkg/events"
)

type ConsumerConfig struct {
	BootstrapServers string
	GroupID          string
	Topics           []string
}

// Generic event handler function type
type EventHandler func(json.RawMessage) error

type KafkaConsumer struct {
	config   ConsumerConfig
	consumer *kafka.Consumer
	handlers map[string]EventHandler
}

// NewKafkaConsumer creates a new KafkaConsumer instance
func NewKafkaConsumer(cfg ConsumerConfig) (*KafkaConsumer, error) {

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":     cfg.BootstrapServers,
		"group.id":              cfg.GroupID,
		"auto.offset.reset":     "earliest",
		"broker.address.family": "v4",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer: %w", err)
	}

	consumer := &KafkaConsumer{
		config:   cfg,
		consumer: c,
		handlers: make(map[string]EventHandler),
	}

	return consumer, nil
}

// RegisterHandler binds a handler to an event type
func (c *KafkaConsumer) RegisterHandler(eventType string, handler EventHandler) {
	c.handlers[eventType] = handler
}

func (c *KafkaConsumer) Start(ctx context.Context) {
	if err := c.consumer.SubscribeTopics(c.config.Topics, nil); err != nil {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := c.consumer.ReadMessage(100) // 100ms
			if err != nil {
				var kafkaErr kafka.Error
				if errors.As(err, &kafkaErr) && kafkaErr.Code() == kafka.ErrTimedOut {
					continue
				}
				continue
			}

			c.handleMessage(msg)
		}
	}
}

func (c *KafkaConsumer) Close() {
	_ = c.consumer.Close()
}

func (c *KafkaConsumer) handleMessage(msg *kafka.Message) {
	var event events.Event
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return
	}

	if handler, ok := c.handlers[event.Type]; ok {
		_ = handler(event.Data)
	}
}
