package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	rdb *redis.Client
}

func New(addr string) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &Client{rdb: rdb}, nil
}

func (c *Client) Close() error {
	return c.rdb.Close()
}

// Account management
func (c *Client) AddAccount(ctx context.Context, username string) error {
	return c.rdb.SAdd(ctx, "accounts", username).Err()
}

func (c *Client) RemoveAccount(ctx context.Context, username string) error {
	return c.rdb.SRem(ctx, "accounts", username).Err()
}

func (c *Client) GetAccounts(ctx context.Context) ([]string, error) {
	return c.rdb.SMembers(ctx, "accounts").Result()
}

func (c *Client) AccountExists(ctx context.Context, username string) (bool, error) {
	return c.rdb.SIsMember(ctx, "accounts", username).Result()
}

// Task queue
func (c *Client) PushTask(ctx context.Context, queue, task string) error {
	return c.rdb.LPush(ctx, queue, task).Err()
}

func (c *Client) PopTask(ctx context.Context, queue string) (string, error) {
	result, err := c.rdb.RPop(ctx, queue).Result()
	if err == redis.Nil {
		return "", nil
	}
	return result, err
}
