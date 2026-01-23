package cache

import (
	"context"
	"errors"
	"fmt"
	"time"
)

const NoExpiration time.Duration = 0

var (
	ErrNotExists      = fmt.Errorf("key not exists")
	ErrScanValueIsNil = fmt.Errorf("scan value is nil")
)

type Interface interface {
	Set(ctx context.Context, key string, value interface{}, expire time.Duration) error
	Update(ctx context.Context, key string, value interface{}) error
	Get(ctx context.Context, key string, value interface{}) error
	Exist(ctx context.Context, key string) (bool, error)
	Remove(ctx context.Context, key string) error
	RemoveWithPattern(ctx context.Context, pattern string) error
	Expire(ctx context.Context, key string, expire time.Duration) error
}

func IsNotExists(e error) bool {
	return errors.Is(e, ErrNotExists)
}
