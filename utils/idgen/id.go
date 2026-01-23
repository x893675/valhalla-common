package idgen

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/sony/sonyflake"
)

var (
	_sf   *sonyflake.Sonyflake
	_once sync.Once
)

// Initialize 初始化 ID 生成器，可选配置
// 如果不调用此函数，将使用默认配置
func Initialize(settings sonyflake.Settings) {
	_once.Do(func() {
		_sf = sonyflake.NewSonyflake(settings)
		if _sf == nil {
			panic("failed to initialize sonyflake")
		}
	})
}

// getSonyflake 获取或初始化 sonyflake 实例
func getSonyflake() *sonyflake.Sonyflake {
	if _sf == nil {
		Initialize(sonyflake.Settings{})
	}
	return _sf
}

// NextID 生成下一个唯一 ID
func NextID() (uint64, error) {
	return getSonyflake().NextID()
}

// MustNextID 生成下一个唯一 ID，出错时 panic
func MustNextID() uint64 {
	id, err := NextID()
	if err != nil {
		panic(fmt.Errorf("failed to generate ID: %w", err))
	}
	return id
}

// NextIDString 生成下一个唯一 ID 的字符串形式
func NextIDString() (string, error) {
	id, err := NextID()
	if err != nil {
		return "", err
	}
	return strconv.FormatUint(id, 10), nil
}

// MustNextIDString 生成下一个唯一 ID 的字符串形式，出错时 panic
func MustNextIDString() string {
	id, err := NextIDString()
	if err != nil {
		panic(fmt.Errorf("failed to generate ID string: %w", err))
	}
	return id
}

// NextIDStringWithPrefix 生成带前缀的 ID 字符串
func NextIDStringWithPrefix(prefix string) (string, error) {
	id, err := NextIDString()
	if err != nil {
		return "", err
	}
	if prefix == "" {
		return id, nil
	}
	return fmt.Sprintf("%s:%s", prefix, id), nil
}

// MustNextIDStringWithPrefix 生成带前缀的 ID 字符串，出错时 panic
func MustNextIDStringWithPrefix(prefix string) string {
	id, err := NextIDStringWithPrefix(prefix)
	if err != nil {
		panic(fmt.Errorf("failed to generate ID string with prefix: %w", err))
	}
	return id
}
