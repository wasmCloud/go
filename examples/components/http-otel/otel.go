package main

import (
	"context"
	"encoding/json"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	witTypes "go.bytecodealliance.org/pkg/wit/types"

	"github.com/wasmCloud/go/examples/components/http-otel/wasi_clocks_wall_clock"
	"github.com/wasmCloud/go/examples/components/http-otel/wasi_otel_tracing"
	"github.com/wasmCloud/go/examples/components/http-otel/wasi_otel_types"
)

// setupOTelSDK configures the global OpenTelemetry SDK to forward span
// lifecycle events to the wasmCloud host through the wasi:otel/tracing
// interface. The host takes care of batching and OTLP export, so no
// exporter or batch processor is needed on the guest side.
func setupOTelSDK() error {
	otel.SetTextMapPropagator(newPropagator())

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(newWasiOtelProcessor()),
		sdktrace.WithResource(sdkresource.NewSchemaless(
			attribute.String("service.name", serviceName),
		)),
	)
	otel.SetTracerProvider(tp)

	return nil
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

// parentFromHost returns ctx with the host's current span context attached
// as a remote parent, so spans started by this component show up as
// children of the host's trace for the incoming request.
func parentFromHost(ctx context.Context) context.Context {
	host := wasi_otel_tracing.OuterSpanContext()

	traceID, err := trace.TraceIDFromHex(host.TraceId)
	if err != nil {
		return ctx
	}
	spanID, err := trace.SpanIDFromHex(host.SpanId)
	if err != nil {
		return ctx
	}

	var flags trace.TraceFlags
	if host.TraceFlags&wasi_otel_tracing.TraceFlagsSampled != 0 {
		flags = flags.WithSampled(true)
	}

	traceState := trace.TraceState{}
	for _, entry := range host.TraceState {
		if ts, err := traceState.Insert(entry.F0, entry.F1); err == nil {
			traceState = ts
		}
	}

	return trace.ContextWithRemoteSpanContext(ctx, trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: flags,
		TraceState: traceState,
		Remote:     true,
	}))
}

// wasiOtelProcessor is a SpanProcessor that mirrors span lifecycle events
// to the wasmCloud host's wasi:otel plugin.
type wasiOtelProcessor struct{}

var _ sdktrace.SpanProcessor = wasiOtelProcessor{}

func newWasiOtelProcessor() sdktrace.SpanProcessor {
	return wasiOtelProcessor{}
}

func (wasiOtelProcessor) OnStart(_ context.Context, s sdktrace.ReadWriteSpan) {
	wasi_otel_tracing.OnStart(toWitSpanContext(s.SpanContext()))
}

func (wasiOtelProcessor) OnEnd(s sdktrace.ReadOnlySpan) {
	wasi_otel_tracing.OnEnd(toWitSpanData(s))
}

func (wasiOtelProcessor) Shutdown(context.Context) error   { return nil }
func (wasiOtelProcessor) ForceFlush(context.Context) error { return nil }

func toWitSpanContext(sc trace.SpanContext) wasi_otel_tracing.SpanContext {
	var traceState []witTypes.Tuple2[string, string]
	sc.TraceState().Walk(func(key, value string) bool {
		traceState = append(traceState, witTypes.Tuple2[string, string]{F0: key, F1: value})
		return true
	})

	return wasi_otel_tracing.SpanContext{
		// trace-id and span-id are hex strings in wasi:otel.
		TraceId:    sc.TraceID().String(),
		SpanId:     sc.SpanID().String(),
		TraceFlags: wasi_otel_tracing.TraceFlags(sc.TraceFlags()) & wasi_otel_tracing.TraceFlagsSampled,
		IsRemote:   sc.IsRemote(),
		TraceState: traceState,
	}
}

func toWitSpanData(s sdktrace.ReadOnlySpan) wasi_otel_tracing.SpanData {
	events := make([]wasi_otel_tracing.Event, 0, len(s.Events()))
	for _, e := range s.Events() {
		events = append(events, wasi_otel_tracing.Event{
			Name:       e.Name,
			Time:       toWitDatetime(e.Time),
			Attributes: toWitAttributes(e.Attributes),
		})
	}

	links := make([]wasi_otel_tracing.Link, 0, len(s.Links()))
	for _, l := range s.Links() {
		links = append(links, wasi_otel_tracing.Link{
			SpanContext: toWitSpanContext(l.SpanContext),
			Attributes:  toWitAttributes(l.Attributes),
		})
	}

	scope := s.InstrumentationScope()

	return wasi_otel_tracing.SpanData{
		SpanContext: toWitSpanContext(s.SpanContext()),
		// All-zero hex string when the span has no parent.
		ParentSpanId:         s.Parent().SpanID().String(),
		SpanKind:             toWitSpanKind(s.SpanKind()),
		Name:                 s.Name(),
		StartTime:            toWitDatetime(s.StartTime()),
		EndTime:              toWitDatetime(s.EndTime()),
		Attributes:           toWitAttributes(s.Attributes()),
		Events:               events,
		Links:                links,
		Status:               toWitStatus(s.Status()),
		InstrumentationScope: toWitScope(scope.Name, scope.Version, scope.SchemaURL, scope.Attributes.ToSlice()),
		DroppedAttributes:    uint32(s.DroppedAttributes()),
		DroppedEvents:        uint32(s.DroppedEvents()),
		DroppedLinks:         uint32(s.DroppedLinks()),
	}
}

func toWitSpanKind(kind trace.SpanKind) wasi_otel_tracing.SpanKind {
	switch kind {
	case trace.SpanKindClient:
		return wasi_otel_tracing.SpanKindClient
	case trace.SpanKindServer:
		return wasi_otel_tracing.SpanKindServer
	case trace.SpanKindProducer:
		return wasi_otel_tracing.SpanKindProducer
	case trace.SpanKindConsumer:
		return wasi_otel_tracing.SpanKindConsumer
	default:
		return wasi_otel_tracing.SpanKindInternal
	}
}

func toWitStatus(status sdktrace.Status) wasi_otel_tracing.Status {
	switch status.Code {
	case codes.Ok:
		return wasi_otel_tracing.MakeStatusOk()
	case codes.Error:
		return wasi_otel_tracing.MakeStatusError(status.Description)
	default:
		return wasi_otel_tracing.MakeStatusUnset()
	}
}

func toWitScope(name string, version, schemaURL string, attrs []attribute.KeyValue) wasi_otel_types.InstrumentationScope {
	scope := wasi_otel_types.InstrumentationScope{
		Name:       name,
		Version:    witTypes.None[string](),
		SchemaUrl:  witTypes.None[string](),
		Attributes: toWitAttributes(attrs),
	}
	if version != "" {
		scope.Version = witTypes.Some(version)
	}
	if schemaURL != "" {
		scope.SchemaUrl = witTypes.Some(schemaURL)
	}
	return scope
}

func toWitDatetime(t time.Time) wasi_clocks_wall_clock.Datetime {
	if t.IsZero() || t.Unix() < 0 {
		return wasi_clocks_wall_clock.Datetime{}
	}
	return wasi_clocks_wall_clock.Datetime{
		Seconds:     uint64(t.Unix()),
		Nanoseconds: uint32(t.Nanosecond()),
	}
}

func toWitAttributes(attrs []attribute.KeyValue) []wasi_otel_types.KeyValue {
	out := make([]wasi_otel_types.KeyValue, 0, len(attrs))
	for _, kv := range attrs {
		out = append(out, wasi_otel_types.KeyValue{
			Key:   string(kv.Key),
			Value: toWitValue(kv.Value),
		})
	}
	return out
}

// toWitValue encodes an attribute value in the wasi:otel `value` format:
// a JSON-serialized AnyValue (WIT has no recursive types, so values are
// carried as JSON strings and decoded by the host).
func toWitValue(v attribute.Value) wasi_otel_types.Value {
	var data any
	switch v.Type() {
	case attribute.BOOL:
		data = v.AsBool()
	case attribute.INT64:
		data = v.AsInt64()
	case attribute.FLOAT64:
		data = v.AsFloat64()
	case attribute.STRING:
		data = v.AsString()
	case attribute.BOOLSLICE:
		data = v.AsBoolSlice()
	case attribute.INT64SLICE:
		data = v.AsInt64Slice()
	case attribute.FLOAT64SLICE:
		data = v.AsFloat64Slice()
	case attribute.STRINGSLICE:
		data = v.AsStringSlice()
	default:
		data = v.Emit()
	}

	encoded, err := json.Marshal(data)
	if err != nil {
		return "null"
	}
	return string(encoded)
}
