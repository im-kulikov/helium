package redis

import (
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

type (
	// Config alias
	Config = redis.Options

	// Client alias
	Client = redis.Client
)

var (
	// ErrEmptyConfig when given empty options
	ErrEmptyConfig = errors.New("redis empty config")
)

// New redis client
func New(opts *Config) (cache *Client, err error) {
	if opts == nil {
		return nil, ErrEmptyConfig
	}

	cache = redis.NewClient(opts)

	if _, err = cache.Ping().Result(); err != nil {
		return nil, err
	}

	return cache, nil
}
