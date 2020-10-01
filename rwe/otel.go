package rwe

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/uptrace/uptrace-go/uptrace"
	"go.opentelemetry.io/otel/api/trace"
)

var (
	upclient *uptrace.Client
	Tracer   trace.Tracer
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

		Resource: map[string]interface{}{
			"service.name": Config.Service,
			"hostname":     hostname,
		},
	})

	Tracer = upclient.Tracer("rwe")

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
