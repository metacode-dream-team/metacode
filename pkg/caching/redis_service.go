package caching

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type RedisService struct {
	client *redis.Client
}

// NewRedisService initializes Redis client from config
func NewRedisService(cfg RedisConfig) (*RedisService, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisService{client: client}, nil
}

func (c *RedisService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

func (c *RedisService) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

func (c *RedisService) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func (c *RedisService) Publish(ctx context.Context, channel, message string) error {
	return c.client.Publish(ctx, channel, message).Err()
}

func (c *RedisService) Exists(ctx context.Context, key string) (bool, error) {
	res, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return res > 0, nil
}

func (c *RedisService) Subscribe(ctx context.Context, channel string) *redis.PubSub {
	return c.client.Subscribe(ctx, channel)
}

// Close gracefully closes Redis connection
func (c *RedisService) Close() error {
	return c.client.Close()
}
