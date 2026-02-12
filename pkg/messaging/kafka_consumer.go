package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/metacode-dream-team/MetaCode/pkg/events"
)

type EventHandler func(json.RawMessage) error

type ConsumerConfig struct {
	BootstrapServers string
	GroupID          string
	Topics           []string
	// EnableLogging toggles our custom logs
	EnableLogging bool
	// LogOutput defines where to send regular logs (e.g., os.Stdout)
	LogOutput io.Writer
	// ErrOutput defines where to send error logs (e.g., os.Stderr)
	ErrOutput io.Writer
}

type KafkaConsumer struct {
	config    ConsumerConfig
	consumer  *kafka.Consumer
	handlers  map[string]EventHandler
	logger    *log.Logger
	errLogger *log.Logger
}

func NewKafkaConsumer(cfg ConsumerConfig) (*KafkaConsumer, error) {
	// Fallback to default writers if not provided
	if cfg.LogOutput == nil {
		cfg.LogOutput = os.Stdout
	}
	if cfg.ErrOutput == nil {
		cfg.ErrOutput = os.Stderr
	}

	kafkaConfig := &kafka.ConfigMap{
		"bootstrap.servers":     cfg.BootstrapServers,
		"group.id":              cfg.GroupID,
		"auto.offset.reset":     "earliest",
		"broker.address.family": "v4",
	}

	// If logging is disabled, we tell librdkafka to be quiet
	if !cfg.EnableLogging {
		kafkaConfig.SetKey("log_level", 0)
	}

	c, err := kafka.NewConsumer(kafkaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer: %w", err)
	}

	return &KafkaConsumer{
		config:    cfg,
		consumer:  c,
		handlers:  make(map[string]EventHandler),
		logger:    log.New(cfg.LogOutput, "[KAFKA_CONSUMER] INFO: ", log.LstdFlags),
		errLogger: log.New(cfg.ErrOutput, "[KAFKA_CONSUMER] ERROR: ", log.LstdFlags),
	}, nil
}

func (c *KafkaConsumer) RegisterHandler(eventType string, handler EventHandler) {
	c.handlers[eventType] = handler
}

func (c *KafkaConsumer) Start(ctx context.Context) {
	if err := c.consumer.SubscribeTopics(c.config.Topics, nil); err != nil {
		c.logErr("Failed to subscribe to topics: %v", err)
		return
	}

	c.logInfo("Consumer started. Subscribed to: %v", c.config.Topics)

	for {
		select {
		case <-ctx.Done():
			c.logInfo("Context cancelled, stopping consumer...")
			return
		default:
			msg, err := c.consumer.ReadMessage(100)
			if err != nil {
				var kafkaErr kafka.Error
				if errors.As(err, &kafkaErr) && kafkaErr.Code() == kafka.ErrTimedOut {
					continue
				}
				c.logErr("Message read error: %v", err)
				continue
			}
			c.handleMessage(msg)
		}
	}
}

func (c *KafkaConsumer) handleMessage(msg *kafka.Message) {
	var event events.Event

	// Logging raw message for debugging if enabled
	c.logInfo("Received message from topic %s", *msg.TopicPartition.Topic)

	if err := json.Unmarshal(msg.Value, &event); err != nil {
		c.logErr("Failed to unmarshal event: %v | Raw: %s", err, string(msg.Value))
		return
	}

	handler, ok := c.handlers[event.Type]
	if !ok {
		c.logErr("No handler registered for event type: %s", event.Type)
		return
	}

	if err := handler(event.Data); err != nil {
		c.logErr("Handler failed for event %s: %v", event.Type, err)
	} else {
		c.logInfo("Successfully processed event: %s", event.Type)
	}
}

// Internal helper for info logging
func (c *KafkaConsumer) logInfo(format string, v ...interface{}) {
	if c.config.EnableLogging {
		c.logger.Printf(format, v...)
	}
}

// Internal helper for error logging
func (c *KafkaConsumer) logErr(format string, v ...interface{}) {
	if c.config.EnableLogging {
		c.errLogger.Printf(format, v...)
	}
}

func (c *KafkaConsumer) Close() {
	_ = c.consumer.Close()
}
