package cache

import (
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
	"server/pkg/failure"
	"time"
)

type TokenCacheAdapter struct {
	client *redis.Client
}

func NewTokenCacheAdapter(client *redis.Client) *TokenCacheAdapter {
	return &TokenCacheAdapter{client}
}

func (c *TokenCacheAdapter) Store(ctx context.Context, token string, ttl time.Duration) error {
	if err := c.client.Set(ctx, token, 1, ttl).Err(); err != nil {
		return failure.NewInternalError("set token error: " + err.Error())
	}
	return nil
}

func (c *TokenCacheAdapter) Check(ctx context.Context, token string) (bool, error) {
	ttl := c.client.TTL(ctx, token)
	if err := ttl.Err(); err != nil {
		return false, failure.NewInternalError("check token error: " + err.Error())
	}
	return ttl.Val() != 0, nil
}

func (c *TokenCacheAdapter) Del(ctx context.Context, token string) error {
	if err := c.client.Del(ctx, token).Err(); err != nil {
		return failure.NewInternalError("delete token error: " + err.Error())
	}
	return nil
}
