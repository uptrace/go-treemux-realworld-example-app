package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	_ "github.com/uptrace/go-realworld-example-app/blog"
	"github.com/uptrace/go-realworld-example-app/httputil"
	_ "github.com/uptrace/go-realworld-example-app/org"
	"github.com/uptrace/go-realworld-example-app/rwe"
	"github.com/uptrace/go-realworld-example-app/xconfig"
)

var listenFlag = flag.String("listen", ":8000", "listen address")

func main() {
	flag.Parse()

	ctx := context.Background()

	cfg, err := xconfig.LoadConfig("api")
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Fatal("LoadConfig failed")
	}

	ctx = rwe.Init(ctx, cfg)
	defer rwe.Exit(ctx)

	var handler http.Handler
	handler = rwe.Router
	handler = httputil.PanicHandler{Next: handler}

	logrus.WithContext(ctx).
		WithField("env", cfg.Env).
		WithField("addr", *listenFlag).
		Info("serving...")

	serveHTTP(ctx, handler)
}

func serveHTTP(ctx context.Context, handler http.Handler) {
	srv := &http.Server{
		Addr:         *listenFlag,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
		Handler:      handler,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && !isServerClosed(err) {
			logrus.WithContext(ctx).WithError(err).Error("ListenAndServe failed")
		}
	}()

	fmt.Println(rwe.WaitExitSignal())

	if err := srv.Shutdown(ctx); err != nil {
		logrus.WithContext(ctx).WithError(err).Error("srv.Shutdown failed")
	}
}

//------------------------------------------------------------------------------

func isServerClosed(err error) bool {
	return err.Error() == "http: Server closed"
}
