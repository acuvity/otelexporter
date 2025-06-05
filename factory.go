// Copyright 2025 Acuvity, Inc.
// SPDX-License-Identifier: Apache-2.0

package otelexporter

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"os"

	"github.com/acuvity/otelexporter/internal/metadata"

	"go.acuvity.ai/a3s/pkgs/token"
	"go.acuvity.ai/manipulate"
	"go.acuvity.ai/manipulate/maniphttp"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/exporter/xexporter"
	"go.opentelemetry.io/collector/pdata/ptrace"

	"go.uber.org/zap"
)

// AcuvityExporter is the interface that must be implemented by the Acuvity exporter.
type AcuvityExporter interface {
	component.Component
	consumeTraces(_ context.Context, td ptrace.Traces) error
}

// NewFactory creates a factory for OTLP exporter.
func NewFactory() exporter.Factory {
	return xexporter.NewFactory(
		metadata.Type,
		createDefaultConfig,
		xexporter.WithTraces(createTracesExporter, metadata.TracesStability),
	)
}

func createDefaultConfig() component.Config {
	return &Config{}
}

func createTracesExporter(
	ctx context.Context,
	set exporter.Settings,
	cfg component.Config,
) (exporter.Traces, error) {
	ae, err := newAcuvityExporter(ctx, cfg, set.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create acuvity exporter: %w", err)
	}
	return exporterhelper.NewTraces(
		ctx,
		set,
		cfg,
		ae.consumeTraces,
		exporterhelper.WithStart(ae.Start),
		exporterhelper.WithShutdown(ae.Shutdown),
		exporterhelper.WithCapabilities(consumer.Capabilities{MutatesData: false}),
	)
}

func newAcuvityExporter(ctx context.Context, cfg component.Config, _ *zap.Logger) (AcuvityExporter, error) {
	conf, ok := cfg.(*Config)
	if !ok {
		return nil, fmt.Errorf("invalid config type: %T", cfg)
	}

	m, err := makeAPIManipulator(ctx, conf)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to backend: %w", err)
	}

	return &acuvityExporter{
		m: m,
	}, nil
}

// makeAPIManipulator prepares a manipulate.Manipulator ready to use with the acuvity backend.
func makeAPIManipulator(ctx context.Context, cfg *Config) (manipulate.Manipulator, error) {

	var identityToken *token.IdentityToken
	var err error
	identityToken, err = token.ParseUnverified(cfg.APIToken)
	if err != nil {
		return nil, fmt.Errorf("unable to parse API token: %w", err)
	}

	// If we have no api url, we infer the API url and eventually namespace based on the api token.
	if cfg.APIURL == "" && cfg.APIToken != "" {
		cfg.APIURL = identityToken.Issuer
		if identityToken.Restrictions.Namespace != "" && cfg.APINamespace == "" {
			cfg.APINamespace = identityToken.Restrictions.Namespace
		}
	}

	systemCAPool, err := systemCAPool(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to get systemCAPool: %w", err)
	}

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS13,
		RootCAs:    systemCAPool,
	}

	opts := []maniphttp.Option{
		maniphttp.OptionNamespace(cfg.APINamespace),
		maniphttp.OptionToken(cfg.APIToken),
		maniphttp.OptionTLSConfig(tlsConfig),
		maniphttp.OptionDefaultRetryFunc(func(i manipulate.RetryInfo) error {
			info := i.(maniphttp.RetryInfo)
			slog.Debug("API manipulator retry",
				"try", info.Try(),
				"method", info.Method,
				"url", info.URL,
				info.Err(),
			)
			return nil
		}),
	}

	m, err := maniphttp.New(ctx, cfg.APIURL, opts...)
	if err != nil {
		return nil, fmt.Errorf("unable to create http manipulator: api-url:%s api-namespace:%s: %w", cfg.APIURL, cfg.APINamespace, err)
	}

	return m, nil
}

// SystemCAPool will return a ca pool either from the system pool
// or a new once using the provided ca flag.
func systemCAPool(c *Config) (*x509.CertPool, error) {

	if c.APICA == "" {
		return x509.SystemCertPool()
	}

	data, err := os.ReadFile(c.APICA)
	if err != nil {
		return nil, fmt.Errorf("unable to read CA file: %w", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(data) {
		return nil, fmt.Errorf("unable to append system signing ca")
	}

	return certPool, nil
}
