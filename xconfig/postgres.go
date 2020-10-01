package xconfig

import (
	"time"

	"github.com/go-pg/pg/v10"
)

type Postgres struct {
	Addr     string
	Database string
	User     string
	Password string
	SSL      bool

	MaxRetries int `yaml:"max_retries"`

	DialTimeout  time.Duration `yaml:"dial_timeout"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`

	PoolSize     int           `yaml:"pool_size"`
	MinIdleConns int           `yaml:"min_idle_conns"`
	MaxConnAge   time.Duration `yaml:"max_conn_age"`
	PoolTimeout  time.Duration `yaml:"pool_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`

	ConnectionPoolPort string `yaml:"connection_pool_port"`
}

func (cfg *Postgres) Options() *pg.Options {
	return &pg.Options{
		User:     cfg.User,
		Password: cfg.Password,
		Database: cfg.Database,

		MaxRetries: cfg.MaxRetries,

		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,

		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		MaxConnAge:   cfg.MaxConnAge,
		PoolTimeout:  cfg.PoolTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
}
