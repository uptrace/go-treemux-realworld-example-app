package rwe

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/uptrace/uptrace-go/uptrace"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/sdk/resource"
)

var (
	upclient *uptrace.Client
	Tracer   = global.Tracer("github.com/uptrace/go-treemux-realworld-example-app")
)

func setupOtel(ctx context.Context) {
	if err := setupUptrace(ctx); err != nil {
		logrus.WithContext(ctx).WithError(err).Error("setupUptrace")
	}
}

func setupUptrace(ctx context.Context) error {
	hostname, _ := os.Hostname()
	upclient = uptrace.NewClient(&uptrace.Config{
		DSN: Config.Uptrace.DSN,

		Resource: resource.New(
			label.String("service.name", Config.Service),
			label.String("host.name", hostname),
		),
	})

	OnExitSecondary(func(ctx context.Context) {
		if err := upclient.Close(); err != nil {
			logrus.WithContext(ctx).WithError(err).Error("uptrace.Close failed")
		}
	})

	return nil
}

func Uptrace() *uptrace.Client {
	return upclient
}
