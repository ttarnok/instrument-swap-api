package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// Global context value for redis.
var ctx = context.Background()

// BlacklistRedisClient handles blacklisting functionality with Redis.
type BlacklistRedisClient struct {
	Client *redis.Client
}

// NewBlacklistRedisClient creates a new BlacklistRedisClient value.
func NewBlacklistRedisClient(addr string, password string, db int) (*BlacklistRedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return &BlacklistRedisClient{}, fmt.Errorf("could not connect to Redis: %w", err)
	}

	return &BlacklistRedisClient{Client: rdb}, nil
}

// BlacklistToken blacklists a given token with the specified expiration time.
func (r *BlacklistRedisClient) BlacklistToken(token string, expiration time.Duration) error {
	err := r.Client.Set(ctx, token, "blacklisted", expiration).Err()
	if err != nil {
		return err
	}
	return nil
}

// IsTokenBlacklisted checks whether the given token is blacklisted.
func (r *BlacklistRedisClient) IsTokenBlacklisted(token string) (bool, error) {
	val, err := r.Client.Get(ctx, token).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return val == "blacklisted", nil
}
