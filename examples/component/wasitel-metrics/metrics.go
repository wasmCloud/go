package main

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.wasmcloud.dev/x/wasitel/wasitelmetric"
)

var (
	metricExporter *wasitelmetric.Exporter
	metricReader   *metric.ManualReader
)

func initMetrics() (err error) {
	// Setup metricExporter
	metricExporter, err = wasitelmetric.New()
	if err != nil {
		return err
	}

	// Setup metricReader
	metricReader = metric.NewManualReader()

	// Setup meterProvider
	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metricReader),
	)
	otel.SetMeterProvider(meterProvider)
	return nil
}

func syncMetrics() error {
	rm := &metricdata.ResourceMetrics{}
	err := metricReader.Collect(context.Background(), rm)
	if err != nil {
		return err
	}

	err = metricExporter.Export(context.Background(), rm)
	if err != nil {
		return err
	}
	return nil
}
