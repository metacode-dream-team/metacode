package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/metacode-dream-team/MetaCode/pkg/events"
)

type EventHandler func(json.RawMessage) error

type ConsumerConfig struct {
	BootstrapServers string
	GroupID          string
	Topics           []string
	EnableLogging    bool
	LogOutput        io.Writer
	ErrOutput        io.Writer

	// ReadTimeout defines how long the consumer waits for a message (ms)
	// Default is 1000ms to prevent CPU busy loops
	ReadTimeout int
	// ErrorBackoff defines pause duration after a non-timeout error
	ErrorBackoff time.Duration
}

type KafkaConsumer struct {
	config    ConsumerConfig
	consumer  *kafka.Consumer
	handlers  map[string]EventHandler
	logger    *log.Logger
	errLogger *log.Logger
}

func NewKafkaConsumer(cfg ConsumerConfig) (*KafkaConsumer, error) {
	// Set default values for optional fields
	if cfg.LogOutput == nil {
		cfg.LogOutput = os.Stdout
	}
	if cfg.ErrOutput == nil {
		cfg.ErrOutput = os.Stderr
	}
	if cfg.ReadTimeout <= 0 {
		cfg.ReadTimeout = 1000 // 1 second default
	}
	if cfg.ErrorBackoff <= 0 {
		cfg.ErrorBackoff = 2 * time.Second
	}

	kafkaConfig := &kafka.ConfigMap{
		"bootstrap.servers":     cfg.BootstrapServers,
		"group.id":              cfg.GroupID,
		"auto.offset.reset":     "earliest",
		"broker.address.family": "v4",
	}

	if !cfg.EnableLogging {
		_ = kafkaConfig.SetKey("log_level", 0)
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
		// Check context cancellation to stop the loop
		if ctx.Err() != nil {
			c.logInfo("Context cancelled, stopping consumer...")
			return
		}

		// ReadMessage is a blocking call up to ReadTimeout.
		// Increasing this value drastically reduces CPU usage during idle periods.
		msg, err := c.consumer.ReadMessage(time.Duration(c.config.ReadTimeout) * time.Millisecond)

		if err != nil {
			var kafkaErr kafka.Error
			if errors.As(err, &kafkaErr) && kafkaErr.Code() == kafka.ErrTimedOut {
				// Expected timeout when no messages are available
				continue
			}

			c.logErr("Message read error: %v. Retrying in %v...", err, c.config.ErrorBackoff)

			// Prevent rapid error loops (e.g., broker disconnect) from spiking CPU
			select {
			case <-time.After(c.config.ErrorBackoff):
				continue
			case <-ctx.Done():
				return
			}
		}

		c.handleMessage(msg)
	}
}

func (c *KafkaConsumer) handleMessage(msg *kafka.Message) {
	var event events.Event

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

func (c *KafkaConsumer) logInfo(format string, v ...interface{}) {
	if c.config.EnableLogging {
		c.logger.Printf(format, v...)
	}
}

func (c *KafkaConsumer) logErr(format string, v ...interface{}) {
	if c.config.EnableLogging {
		c.errLogger.Printf(format, v...)
	}
}

func (c *KafkaConsumer) Close() {
	c.logInfo("Closing Kafka consumer...")
	_ = c.consumer.Close()
}
