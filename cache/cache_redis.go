package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	redisv9 "github.com/redis/go-redis/v9"
)

type redisKV struct {
	client redisv9.Cmdable
}

func (r *redisKV) Set(ctx context.Context, key string, value interface{}, expire time.Duration) error {
	_, err := r.client.Set(context.TODO(), key, value, expire).Result()
	return err
}

func (r *redisKV) Update(ctx context.Context, key string, value interface{}) error {
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *redisKV) Get(ctx context.Context, key string, value interface{}) error {
	err := r.client.Get(ctx, key).Scan(value)
	if errors.Is(redisv9.Nil, err) {
		return ErrNotExists
	}
	return err
}

func (r *redisKV) Exist(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	return count > 0, err
}

func (r *redisKV) Remove(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *redisKV) Expire(ctx context.Context, key string, expire time.Duration) error {
	return r.client.Expire(ctx, key, expire).Err()
}

func (r *redisKV) RemoveWithPattern(ctx context.Context, pattern string) error {
	var cursor uint64
	var n int

	for {
		var keys []string
		var err error
		keys, cursor, err = r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}
		n += len(keys)
		if len(keys) > 0 {
			r.client.Del(ctx, keys...)
		}
		if cursor == 0 {
			break
		}
	}
	return nil
}

func NewRedis(opt *RedisOptions) (Interface, error) {
	if len(opt.Addrs) == 0 {
		return nil, fmt.Errorf("redis addresses cannot be empty")
	}

	kv := redisKV{}

	switch opt.Schema {
	case Redis:
		kv.client = redisv9.NewClient(&redisv9.Options{
			Addr:     opt.Addrs[0],
			Username: opt.Username,
			Password: opt.Password,
			DB:       opt.DB,
		})
	case RedisSentinel:
		kv.client = redisv9.NewFailoverClient(&redisv9.FailoverOptions{
			SentinelAddrs: opt.Addrs,
			Username:      opt.Username,
			Password:      opt.Password,
			DB:            opt.DB,
		})
	case RedisCluster:
		kv.client = redisv9.NewClusterClient(&redisv9.ClusterOptions{
			Addrs:    opt.Addrs,
			Username: opt.Username,
			Password: opt.Password,
		})
	default:
		return nil, fmt.Errorf("not support redis schema:%s", opt.Schema)
	}

	return &kv, nil
}
