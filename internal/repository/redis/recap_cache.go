package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Valentin0851/avito-recap-backend/internal/usecase/ports"
	"github.com/redis/go-redis/v9"
)

const RecapTTL = 24 * time.Hour

type RecapCache struct {
	client *redis.Client
}

func NewRecapCache(client *redis.Client) *RecapCache {
	return &RecapCache{client: client}
}

func (c *RecapCache) Get(ctx context.Context, key string, dest any) (bool, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to get from cache: %w", err)
	}
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return false, fmt.Errorf("failed to unmarshal cache value: %w", err)
	}
	return true, nil
}

func (c *RecapCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}
	return nil
}

func (c *RecapCache) Delete(ctx context.Context, key string) error {
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete cache: %w", err)
	}
	return nil
}

var _ ports.Cache = (*RecapCache)(nil)
