package rwe

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/uptrace/uptrace-go/uptrace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/semconv"
)

var (
	upclient *uptrace.Client
	Tracer   = otel.Tracer("github.com/uptrace/go-treemux-realworld-example-app")
)

func setupOtel(ctx context.Context) {
	if err := setupUptrace(ctx); err != nil {
		logrus.WithContext(ctx).WithError(err).Error("setupUptrace")
	}
}

func setupUptrace(ctx context.Context) error {
	resource, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(Config.Service),
		),
	)
	if err != nil {
		panic(err)
	}

	upclient = uptrace.NewClient(&uptrace.Config{
		DSN:      Config.Uptrace.DSN,
		Resource: resource,
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
