package resource

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const (
	defaultDialTimeout  = 2 * time.Second
	defaultReadTimeout  = time.Second
	defaultWriteTimeout = time.Second
	defaultPoolSize     = 32
)

// Config contains shared Redis client and pool settings.
type RedisConfig struct {
	Address      string
	Username     string
	Password     string
	DB           int
	PoolSize     int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// Client is the common Redis capability used by all node types. Business key
// formats and scripts belong to their owning service, not this package.
type RedisClient struct {
	client *goredis.Client
}

func openRedis(ctx context.Context, config RedisConfig) (*RedisClient, error) {
	config = withRedisDefaults(config)
	rdb := goredis.NewClient(&goredis.Options{
		Addr:         config.Address,
		Username:     config.Username,
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return &RedisClient{client: rdb}, nil
}

func (c *RedisClient) Command() goredis.UniversalClient { return c.client }

func (c *RedisClient) Ping(ctx context.Context) error { return c.client.Ping(ctx).Err() }

func (c *RedisClient) Close() error { return c.client.Close() }

func withRedisDefaults(config RedisConfig) RedisConfig {
	if config.PoolSize <= 0 {
		config.PoolSize = defaultPoolSize
	}
	if config.DialTimeout <= 0 {
		config.DialTimeout = defaultDialTimeout
	}
	if config.ReadTimeout <= 0 {
		config.ReadTimeout = defaultReadTimeout
	}
	if config.WriteTimeout <= 0 {
		config.WriteTimeout = defaultWriteTimeout
	}
	return config
}
