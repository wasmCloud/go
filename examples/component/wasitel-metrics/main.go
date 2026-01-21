//go:generate go tool wit-bindgen-go generate --world metrics --out gen ./wit

package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.wasmcloud.dev/component/log/wasilog"
	"go.wasmcloud.dev/component/net/wasihttp"
)

var logger = wasilog.DefaultLogger

var meter = otel.Meter("wasitel-metrics")

func init() {
	if err := initMetrics(); err != nil {
		logger.Error("Failed to init metrics", "error", err)
	}

	router := http.NewServeMux()
	router.HandleFunc("/", metricsMiddleware(echoHandler))
	wasihttp.Handle(router)
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := syncMetrics(); err != nil {
			logger.Error("Failed to sync metrics", "error", err)
		}
	}()

	fmt.Fprintf(w, "Hello Metrics!")
}

func metricsMiddleware(next func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	requestCount, err := meter.Int64UpDownCounter("request_count")
	if err != nil {
		logger.Error("failed to create counter", "error", err)
	}

	responseTime, err := meter.Float64Histogram("response_time")
	if err != nil {
		logger.Error("failed to create histogram", "error", err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := syncMetrics(); err != nil {
				logger.Error("Failed to sync metrics", "error", err)
			}
		}()
		reqId := uuid.NewString()

		requestCount.Add(r.Context(), 1, metric.WithAttributes(
			attribute.String("request_id", reqId),
		))

		startTime := time.Now()
		next(w, r)
		endTime := time.Now()

		responseTime.Record(r.Context(), endTime.Sub(startTime).Seconds(), metric.WithAttributes(
			attribute.String("request_id", reqId),
		))
	}
}

func main() {}
