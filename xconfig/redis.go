package xconfig

import (
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisRing struct {
	Addrs    map[string]string
	Password string
	DB       int

	MaxRetries int `yaml:"max_retries"`

	PoolSize    int           `yaml:"pool_size"`
	PoolTimeout time.Duration `yaml:"pool_timeout"`
	IdleTimeout time.Duration `yaml:"idle_timeout"`

	DialTimeout  time.Duration `yaml:"dial_timeout"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

func (cfg *RedisRing) Options() *redis.RingOptions {
	return &redis.RingOptions{
		Addrs:    cfg.Addrs,
		Password: cfg.Password,
		DB:       cfg.DB,

		MaxRetries: cfg.MaxRetries,

		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,

		PoolSize:    cfg.PoolSize,
		PoolTimeout: cfg.PoolTimeout,
		IdleTimeout: cfg.IdleTimeout,
	}
}
