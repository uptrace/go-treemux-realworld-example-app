package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/go-pg/migrations/v8"
	"github.com/sirupsen/logrus"
	"github.com/uptrace/go-realworld-example-app/rwe"
	"github.com/uptrace/go-realworld-example-app/xconfig"
)

const stmtTimeout = 5 * time.Minute

func main() {
	flag.Parse()

	cfg, err := xconfig.LoadConfig("migrate_db")
	if err != nil {
		logrus.Fatal(err)
	}

	cfg.PGMain.ReadTimeout = stmtTimeout
	cfg.PGMain.PoolTimeout = stmtTimeout

	ctx := rwe.Init(context.Background(), cfg)
	defer rwe.Exit(ctx)

	args := flag.Args()
	oldVersion, newVersion, err := migrations.Run(rwe.PGMain(), args...)
	if err != nil {
		logrus.Fatalf("migration %d -> %d failed: %s",
			oldVersion, newVersion, err)
	}

	if newVersion != oldVersion {
		fmt.Printf("migrated from %d to %d\n", oldVersion, newVersion)
	} else {
		fmt.Printf("version is %d\n", oldVersion)
	}
}
