package cache

import "fmt"

type Options struct {
	Type  string        `json:"type" yaml:"type" toml:"type"`
	Redis *RedisOptions `json:"redis" yaml:"redis" toml:"redis"`
}

const (
	Redis         = "redis"
	RedisSentinel = "redis-sentinel"
	RedisCluster  = "redis-cluster"
)

type RedisOptions struct {
	// redis schema. one of redis redis-sentinel cluster
	Schema string `json:"schema" yaml:"schema" toml:"schema"`

	Addrs    []string `json:"addrs" yaml:"addrs" toml:"addrs"`
	Username string   `json:"username" yaml:"username" toml:"username"`
	Password string   `json:"password" yaml:"password" toml:"password"`
	DB       int      `json:"db" yaml:"db" toml:"db"`

	MasterName       string `json:"masterName" yaml:"masterName" toml:"masterName"`
	SentinelUsername string `json:"sentinelUsername" yaml:"sentinelUsername" toml:"sentinelUsername"`
	SentinelPassword string `json:"sentinelPassword" yaml:"sentinelPassword" toml:"sentinelPassword"`
}

func DefaultOptions() *Options {
	return &Options{
		Type: "mem",
	}
}

func New(opts *Options) (Interface, error) {
	switch opts.Type {
	case "mem":
		return NewMemory()
	case Redis:
		return NewRedis(opts.Redis)
	default:
		return nil, fmt.Errorf("not support cache type:%s", opts.Type)
	}
}
