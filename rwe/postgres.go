package rwe

import (
	"context"
	"net"
	"sync"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pgext"
	"github.com/sirupsen/logrus"
	"github.com/uptrace/go-realworld-example-app/xconfig"
)

var (
	pgMainOnce sync.Once
	pgMain     *pg.DB
)

func PGMain() *pg.DB {
	pgMainOnce.Do(func() {
		pgMain = NewPostgres(Config.PGMain, hasPgbouncer())
	})
	return pgMain
}

var (
	pgMainTxOnce sync.Once
	pgMainTx     *pg.DB
)

func PGMainTx() *pg.DB {
	pgMainTxOnce.Do(func() {
		pgMainTx = NewPostgres(Config.PGMain, false)
	})
	return pgMainTx
}

func hasPgbouncer() bool {
	switch Config.Env {
	case "test", "dev":
		return false
	default:
		return true
	}
}

func NewPostgres(cfg *xconfig.Postgres, usePool bool) *pg.DB {
	addr := cfg.Addr
	if usePool {
		addr = replacePort(addr, cfg.ConnectionPoolPort)
	}

	opt := cfg.Options()
	opt.Addr = addr

	db := pg.Connect(opt)
	OnExitSecondary(func(ctx context.Context) {
		if err := db.Close(); err != nil {
			logrus.WithError(err).Error("pg.Close failed")
		}
	})

	db.AddQueryHook(pgext.OpenTelemetryHook{})
	if IsDebug() {
		db.AddQueryHook(pgext.DebugHook{})
	}

	return db
}

func replacePort(s, newPort string) string {
	if newPort == "" {
		return s
	}
	host, _, err := net.SplitHostPort(s)
	if err != nil {
		host = s
	}
	return net.JoinHostPort(host, newPort)
}
