// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package telemetry // import "go.opentelemetry.io/collector/service/telemetry"

import (
	"context"
	"errors"
	"net/http"
	"sync"

	"go.opentelemetry.io/contrib/config"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.uber.org/multierr"
	"go.uber.org/zap"

	"go.opentelemetry.io/collector/config/configtelemetry"
	semconv "go.opentelemetry.io/collector/semconv/v1.13.0"
	"go.opentelemetry.io/collector/service/telemetry/internal/otelinit"
)

const (
	zapKeyTelemetryAddress = "address"
	zapKeyTelemetryLevel   = "metrics level"
)

var (

	// gRPC Instrumentation Name
	GRPCInstrumentation = "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	// http Instrumentation Name
	HTTPInstrumentation = "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	// GRPCUnacceptableKeyValues is a list of high cardinality grpc attributes that should be filtered out.
	GRPCUnacceptableKeyValues = []string{
		semconv.AttributeNetSockPeerAddr,
		semconv.AttributeNetSockPeerPort,
		semconv.AttributeNetSockPeerName,
	}

	// HTTPUnacceptableKeyValues is a list of high cardinality http attributes that should be filtered out.
	HTTPUnacceptableKeyValues = []string{
		semconv.AttributeNetHostName,
		semconv.AttributeNetHostPort,
	}
)

type meterProvider struct {
	*sdkmetric.MeterProvider
	servers  []*http.Server
	serverWG sync.WaitGroup
}

type meterProviderSettings struct {
	res               *resource.Resource
	cfg               MetricsConfig
	asyncErrorChannel chan error
}

func newerMeterProvider(set Settings, cfg Config) (metric.MeterProvider, error) {
	if cfg.Metrics.Level == configtelemetry.LevelNone || len(cfg.Metrics.Readers) == 0 {
		return noop.NewMeterProvider(), nil
	}
	if set.SDK != nil {
		return set.SDK.MeterProvider(), nil
	}
	return nil, errors.New("no sdk set")
}

// newMeterProvider creates a new MeterProvider from Config.
func newMeterProvider(set meterProviderSettings, disableHighCardinality bool) (metric.MeterProvider, error) {
	if set.cfg.Level == configtelemetry.LevelNone || len(set.cfg.Readers) == 0 {
		return noop.NewMeterProvider(), nil
	}

	mp := &meterProvider{}
	var opts []sdkmetric.Option
	for _, reader := range set.cfg.Readers {
		// https://github.com/open-telemetry/opentelemetry-collector/issues/8045
		r, server, err := otelinit.InitMetricReader(context.Background(), reader, set.asyncErrorChannel, &mp.serverWG)
		if err != nil {
			return nil, err
		}
		if server != nil {
			mp.servers = append(mp.servers, server)
		}
		opts = append(opts, sdkmetric.WithReader(r))
	}

	var err error
	mp.MeterProvider, err = otelinit.InitOpenTelemetry(set.res, opts, disableHighCardinality)
	if err != nil {
		return nil, err
	}
	return mp, nil
}

// LogAboutServers logs about the servers that are serving metrics.
func (mp *meterProvider) LogAboutServers(logger *zap.Logger, cfg MetricsConfig) {
	for _, server := range mp.servers {
		logger.Info(
			"Serving metrics",
			zap.String(zapKeyTelemetryAddress, server.Addr),
			zap.Stringer(zapKeyTelemetryLevel, cfg.Level),
		)
	}
}

// Shutdown the meter provider and all the associated resources.
// The type signature of this method matches that of the sdkmetric.MeterProvider.
func (mp *meterProvider) Shutdown(ctx context.Context) error {
	var errs error
	for _, server := range mp.servers {
		if server != nil {
			errs = multierr.Append(errs, server.Close())
		}
	}
	errs = multierr.Append(errs, mp.MeterProvider.Shutdown(ctx))
	mp.serverWG.Wait()

	return errs
}

func disableConfigHighCardinalityViews() []config.View {
	return []config.View{
		{
			Selector: &config.ViewSelector{
				InstrumentName: &GRPCInstrumentation,
			},
			Stream: &config.ViewStream{
				AttributeKeys: GRPCUnacceptableKeyValues,
			},
		},
		{
			Selector: &config.ViewSelector{
				InstrumentName: &HTTPInstrumentation,
			},
			Stream: &config.ViewStream{
				AttributeKeys: HTTPUnacceptableKeyValues,
			},
		},
	}
}
