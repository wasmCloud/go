package main

import (
	"time"
	_ "unsafe"

	"go.wasmcloud.dev/x/wasitel/wasiteltrace"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
)

func setupOTelSDK() error {
	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// // Set up trace provider.
	tp, err := newTraceProvider()
	if err != nil {
		return err
	}
	otel.SetTracerProvider(tp)

	return nil
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTraceProvider() (*trace.TracerProvider, error) {
	traceExporter, err := wasiteltrace.New()
	if err != nil {
		return nil, err
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter,
			// Default is 5s. Set to 1s for demonstrative purposes.
			trace.WithBatchTimeout(time.Second)),
	)
	return traceProvider, nil
}
