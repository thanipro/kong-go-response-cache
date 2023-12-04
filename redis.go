package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"time"
)

type RedisClient struct {
	Client *redis.Client
	Ctx    context.Context
}

// NewRedisClient creates a new Redis client using the provided configuration
func NewRedisClient(cfg Config) (*RedisClient, error) {
	ctx := context.Background()
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:        fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password:    cfg.RedisPassword,
		DB:          0,
		IdleTimeout: time.Duration(cfg.CachedTtl) * time.Second,
	})

	return &RedisClient{Client: rdb, Ctx: ctx}, nil
}

// Store stores a value in Redis with the given key
func (rc *RedisClient) Store(key string, value interface{}) error {
	cachedDuration, err := time.ParseDuration(rc.Client.Options().IdleTimeout.String())
	if err != nil {
		return fmt.Errorf("error parsing cached duration: %w", err)
	}

	if err := rc.Client.Set(rc.Ctx, key, value, cachedDuration).Err(); err != nil {
		return fmt.Errorf("error setting value in Redis: %w", err)
	}

	return nil
}

// Retrieve gets a value from Redis by key
func (rc *RedisClient) Retrieve(key string) (string, error) {
	val, err := rc.Client.Get(rc.Ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("error retrieving value from Redis: %w", err)
	}
	return val, nil
}

// validateConfig checks if the configuration is valid
func validateConfig(cfg Config) error {
	if cfg.RedisHost == "" || cfg.RedisPort == "" {
		return fmt.Errorf("redis host and port must be provided")
	}
	if cfg.CachedTtl < 0 {
		return fmt.Errorf("cached TTL must be non-negative")
	}
	return nil
}
