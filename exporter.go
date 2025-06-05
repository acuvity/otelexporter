// Copyright 2025 Acuvity, Inc.
// SPDX-License-Identifier: Apache-2.0

package otelexporter

import (
	"context"

	api "go.acuvity.ai/api/backend"
	"go.acuvity.ai/manipulate"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

// acuvityExporter is the implementation of file exporter that writes telemetry data to a file
type acuvityExporter struct {
	m         manipulate.Manipulator
	tokenName string
}

func (e *acuvityExporter) consumeTraces(ctx context.Context, td ptrace.Traces) error {

	it := api.NewIngestTrace()
	it.Principal.Type = api.PrincipalTypeApp
	it.Principal.AuthType = api.PrincipalAuthTypeAppToken
	it.Principal.TokenName = e.tokenName
	it.Principal.App = api.NewPrincipalApp()
	it.Principal.App.Name = e.tokenName

	for _, resourceSpans := range td.ResourceSpans().All() {
		for _, scopeSpans := range resourceSpans.ScopeSpans().All() {
			for _, span := range scopeSpans.Spans().All() {

				tr := api.NewTraceRef()

				tr.TraceID = span.TraceID().String()
				tr.SpanID = span.SpanID().String()
				tr.ParentSpanID = span.ParentSpanID().String()

				tr.SpanName = span.Name()

				tr.SpanStart = span.StartTimestamp().AsTime()
				tr.SpanEnd = span.EndTimestamp().AsTime()

				switch span.Kind() {
				case ptrace.SpanKindClient:
					tr.Kind = api.TraceRefKindClient
				case ptrace.SpanKindServer:
					tr.Kind = api.TraceRefKindServer
				case ptrace.SpanKindInternal:
					tr.Kind = api.TraceRefKindInternal
				case ptrace.SpanKindProducer:
					tr.Kind = api.TraceRefKindProducer
				case ptrace.SpanKindConsumer:
					tr.Kind = api.TraceRefKindConsumer
				}

				switch span.Status().Code() {
				case ptrace.StatusCodeOk:
					tr.StatusCode = api.TraceRefStatusCodeOK
				case ptrace.StatusCodeError:
					tr.StatusCode = api.TraceRefStatusCodeError
					tr.StatusMessage = span.Status().Message()
				}

				it.Traces = append(it.Traces, tr)
			}
		}
	}

	if len(it.Traces) == 0 {
		return nil
	}

	mctx := manipulate.NewContext(ctx)
	return e.m.Create(mctx, it)
}

// Start starts the flush timer if set.
func (e *acuvityExporter) Start(_ context.Context, host component.Host) error {
	return nil
}

// Shutdown stops the exporter and is invoked during shutdown.
// It stops the flush ticker if set.
func (e *acuvityExporter) Shutdown(context.Context) error {
	return nil
}

// NewAcuvityExporter creates a new Acuvity exporter from the given manipulator.
func NewAcuvityExporterFromManipulator(m manipulate.Manipulator) (consumer.Traces, error) {
	e := &acuvityExporter{
		m: m,
	}
	return consumer.NewTraces(
		e.consumeTraces,
		consumer.WithCapabilities(consumer.Capabilities{MutatesData: false}),
	)
}
