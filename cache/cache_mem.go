package cache

import (
	"context"
	"encoding"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

type entry struct {
	expireAt time.Time
	value    []byte
}

func (e entry) scan(value interface{}) error {
	switch v := value.(type) {
	case nil:
		return fmt.Errorf("memory cache: can't scan %T", v)
	case *string:
		*v = string(e.value)
		return nil
	case *[]byte:
		*v = e.value
		return nil
	case *int:
		var err error
		*v, err = strconv.Atoi(string(e.value))
		return err
	case *int8:
		n, err := strconv.ParseInt(string(e.value), 10, 8)
		if err != nil {
			return err
		}
		*v = int8(n)
		return nil
	case *int16:
		n, err := strconv.ParseInt(string(e.value), 10, 16)
		if err != nil {
			return err
		}
		*v = int16(n)
		return nil
	case *int32:
		n, err := strconv.ParseInt(string(e.value), 10, 32)
		if err != nil {
			return err
		}
		*v = int32(n)
		return nil
	case *int64:
		n, err := strconv.ParseInt(string(e.value), 10, 64)
		if err != nil {
			return err
		}
		*v = n
		return nil
	case *uint:
		n, err := strconv.ParseUint(string(e.value), 10, 64)
		if err != nil {
			return err
		}
		*v = uint(n)
		return nil
	case *uint8:
		n, err := strconv.ParseUint(string(e.value), 10, 8)
		if err != nil {
			return err
		}
		*v = uint8(n)
		return nil
	case *uint16:
		n, err := strconv.ParseUint(string(e.value), 10, 16)
		if err != nil {
			return err
		}
		*v = uint16(n)
		return nil
	case *uint32:
		n, err := strconv.ParseUint(string(e.value), 10, 32)
		if err != nil {
			return err
		}
		*v = uint32(n)
		return nil
	case *uint64:
		n, err := strconv.ParseUint(string(e.value), 10, 64)
		if err != nil {
			return err
		}
		*v = n
		return nil
	case *float32:
		n, err := strconv.ParseFloat(string(e.value), 32)
		if err != nil {
			return err
		}
		*v = float32(n)
		return nil
	case *float64:
		n, err := strconv.ParseFloat(string(e.value), 64)
		if err != nil {
			return err
		}
		*v = n
		return nil
	case *bool:
		n, err := strconv.ParseBool(string(e.value))
		if err != nil {
			return err
		}
		*v = n
		return nil
	case *time.Time:
		var err error
		*v, err = time.Parse(time.RFC3339, string(e.value))
		if err != nil {
			return err
		}
		return nil
	case *time.Duration:
		n, err := strconv.ParseInt(string(e.value), 10, 64)
		if err != nil {
			return err
		}
		*v = time.Duration(n)
		return nil
	case encoding.BinaryUnmarshaler:
		return v.UnmarshalBinary(e.value)
	default:
		return fmt.Errorf("memory cache: can't unmarshall %T (implement json.Unmarshaler)", v)
	}
}

type memoryKV struct {
	storage *sync.Map
	Now     func() time.Time
}

func (m *memoryKV) get(key string) (*entry, error) {
	v, ok := m.storage.Load(key)
	if !ok {
		return nil, ErrNotExists
	}
	e := v.(entry)
	if !e.expireAt.IsZero() {
		now := m.Now()
		if now.After(e.expireAt) {
			m.storage.Delete(key)
			return nil, ErrNotExists
		}
	}
	return &e, nil
}

func (m *memoryKV) Update(ctx context.Context, key string, value interface{}) error {
	e, err := m.get(key)
	if err != nil {
		return err
	}
	e.value, err = marshallValue(value)
	if err != nil {
		return err
	}
	m.storage.Store(key, e)
	return nil
}

func (m *memoryKV) Get(ctx context.Context, key string, value interface{}) error {
	if value == nil {
		return ErrScanValueIsNil
	}
	e, err := m.get(key)
	if err != nil {
		return err
	}
	return e.scan(value)
}

func (m *memoryKV) Exist(ctx context.Context, key string) (bool, error) {
	_, err := m.get(key)
	if err != nil {
		if IsNotExists(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (m *memoryKV) Remove(ctx context.Context, key string) error {
	m.storage.Delete(key)
	return nil
}

func (m *memoryKV) Expire(ctx context.Context, key string, expire time.Duration) error {
	e, err := m.get(key)
	if err != nil {
		return err
	}
	e.expireAt = m.Now().Add(expire)

	m.storage.Store(key, *e)
	return nil
}

func (m *memoryKV) Set(ctx context.Context, key string, value interface{}, expire time.Duration) error {
	var (
		expireAt time.Time
		err      error
	)
	if expire > NoExpiration {
		expireAt = m.Now().Add(expire)
	}
	e := entry{
		expireAt: expireAt,
	}
	e.value, err = marshallValue(value)
	if err != nil {
		return err
	}
	m.storage.Store(key, e)
	return nil
}

func marshallValue(value interface{}) ([]byte, error) {
	switch v := value.(type) {
	case nil:
		return []byte(""), nil
	case string:
		return []byte(v), nil
	case []byte:
		return v, nil
	case int:
		return []byte(strconv.FormatInt(int64(v), 10)), nil
	case int8:
		return []byte(strconv.FormatInt(int64(v), 10)), nil
	case int16:
		return []byte(strconv.FormatInt(int64(v), 10)), nil
	case int32:
		return []byte(strconv.FormatInt(int64(v), 10)), nil
	case int64:
		return []byte(strconv.FormatInt(v, 10)), nil
	case uint:
		return []byte(strconv.FormatUint(uint64(v), 10)), nil
	case uint8:
		return []byte(strconv.FormatUint(uint64(v), 10)), nil
	case uint16:
		return []byte(strconv.FormatUint(uint64(v), 10)), nil
	case uint32:
		return []byte(strconv.FormatUint(uint64(v), 10)), nil
	case uint64:
		return []byte(strconv.FormatUint(v, 10)), nil
	case float32:
		return []byte(strconv.FormatFloat(float64(v), 'f', -1, 32)), nil
	case float64:
		return []byte(strconv.FormatFloat(v, 'f', -1, 64)), nil
	case bool:
		if v {
			return []byte("1"), nil
		}
		return []byte("0"), nil
	case time.Time:
		return []byte(v.Format(time.RFC3339)), nil
	case time.Duration:
		return []byte(strconv.FormatInt(int64(v), 10)), nil
	case encoding.BinaryMarshaler:
		return v.MarshalBinary()
	default:
		return nil, fmt.Errorf(
			"memory cache: can't marshal %T (implement encoding.BinaryMarshaler)", v)
	}
}

// RemoveWithPattern removes all keys with the given pattern.
// memoryKV only support pattern with suffix "*". eg: `prefix:*` will remove all keys with `prefix:`
func (m *memoryKV) RemoveWithPattern(ctx context.Context, pattern string) error {
	var keys []string
	prefix := strings.TrimSuffix(pattern, "*")
	m.storage.Range(func(key, value interface{}) bool {
		k := key.(string)
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
		return true
	})
	for _, k := range keys {
		m.storage.Delete(k)
	}
	return nil
}

func NewMemory() (Interface, error) {
	return &memoryKV{
		storage: &sync.Map{},
		Now:     time.Now,
	}, nil
}
