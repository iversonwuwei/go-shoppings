package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"wechat-mall-saas/internal/pkg/config"
)

type Client struct {
	*redis.Client
}

func New(cfg config.RedisConfig) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return &Client{rdb}, nil
}

// Lock 分布式锁：SETNX + TTL，返回 token 用于解锁
func (c *Client) Lock(ctx context.Context, key, token string, ttl time.Duration) (bool, error) {
	return c.SetNX(ctx, key, token, ttl).Result()
}

// Unlock 仅当 token 匹配时才删除
func (c *Client) Unlock(ctx context.Context, key, token string) error {
	script := `if redis.call("get", KEYS[1]) == ARGV[1] then return redis.call("del", KEYS[1]) else return 0 end`
	return c.Eval(ctx, script, []string{key}, token).Err()
}
