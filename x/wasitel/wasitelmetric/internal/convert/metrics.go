package convert

import (
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.wasmcloud.dev/x/wasitel/wasitelmetric/internal/types"
)

func ResourceMetrics(rm *metricdata.ResourceMetrics) (*types.ResourceMetrics, error) {
	return &types.ResourceMetrics{
		Resource: &types.Resource{
			Attributes: AttrIter(rm.Resource.Iter()),
		},
		ScopeMetrics: ScopeMetrics(rm.ScopeMetrics),
		SchemaUrl:    rm.Resource.SchemaURL(),
	}, nil
}

func ScopeMetrics(sms []metricdata.ScopeMetrics) []*types.ScopeMetrics {
	out := make([]*types.ScopeMetrics, len(sms))
	for i, sm := range sms {
		out[i] = &types.ScopeMetrics{
			Scope: &types.InstrumentationScope{
				Name:       sm.Scope.Name,
				Version:    sm.Scope.Version,
				Attributes: AttrIter(sm.Scope.Attributes.Iter()),
			},
			Metrics:   Metrics(sm.Metrics),
			SchemaUrl: sm.Scope.SchemaURL,
		}
	}
	return out
}

func Metrics(ms []metricdata.Metrics) []*types.Metric {
	out := make([]*types.Metric, len(ms))
	for i, m := range ms {
		out[i] = metric(m)
	}
	return out
}

func metric(m metricdata.Metrics) *types.Metric {
	out := &types.Metric{
		Name:        m.Name,
		Description: m.Description,
		Unit:        m.Unit,
	}
	switch a := m.Data.(type) {
	case metricdata.Gauge[int64]:
		out.Gauge = Gauge(a)
	case metricdata.Gauge[float64]:
		out.Gauge = Gauge(a)
	case metricdata.Sum[int64]:
		out.Sum = Sum(a)
	case metricdata.Sum[float64]:
		out.Sum = Sum(a)
	case metricdata.Histogram[int64]:
		out.Histogram = Histogram(a)
	case metricdata.Histogram[float64]:
		out.Histogram = Histogram(a)
	case metricdata.ExponentialHistogram[int64]:
		out.ExponentialHistogram = ExponentialHistogram(a)
	case metricdata.ExponentialHistogram[float64]:
		out.ExponentialHistogram = ExponentialHistogram(a)
	case metricdata.Summary:
		out.Summary = Summary(a)
	}
	return out
}

func Gauge[N int64 | float64](g metricdata.Gauge[N]) *types.Gauge {
	return &types.Gauge{
		DataPoints: DataPoints(g.DataPoints),
	}
}

func Sum[N int64 | float64](s metricdata.Sum[N]) *types.Sum {
	return &types.Sum{
		AggregationTemporality: types.AggregationTemporality(s.Temporality),
		IsMonotonic:            s.IsMonotonic,
		DataPoints:             DataPoints(s.DataPoints),
	}
}

func DataPoints[N int64 | float64](pts []metricdata.DataPoint[N]) []*types.NumberDataPoint {
	out := make([]*types.NumberDataPoint, len(pts))
	for i, pt := range pts {
		out[i] = &types.NumberDataPoint{
			Attributes:        AttrIter(pt.Attributes.Iter()),
			StartTimeUnixNano: uint64(pt.StartTime.UnixNano()),
			TimeUnixNano:      uint64(pt.Time.UnixNano()),
			Exemplars:         Exemplars(pt.Exemplars),
		}
	}
	return out
}

func Histogram[N int64 | float64](h metricdata.Histogram[N]) *types.Histogram {
	return &types.Histogram{
		AggregationTemporality: types.AggregationTemporality(h.Temporality),
		DataPoints:             HistogramDataPoints(h.DataPoints),
	}
}

func HistogramDataPoints[N int64 | float64](pts []metricdata.HistogramDataPoint[N]) []*types.HistogramDataPoint {
	out := make([]*types.HistogramDataPoint, len(pts))
	for i, pt := range pts {
		sum := float64(pt.Sum)
		hdp := &types.HistogramDataPoint{
			Attributes:        AttrIter(pt.Attributes.Iter()),
			StartTimeUnixNano: uint64(pt.StartTime.UnixNano()),
			TimeUnixNano:      uint64(pt.Time.UnixNano()),
			Count:             pt.Count,
			Sum:               &sum,
			BucketCounts:      pt.BucketCounts,
			ExplicitBounds:    pt.Bounds,
			Exemplars:         Exemplars(pt.Exemplars),
		}
		if v, ok := pt.Min.Value(); ok {
			vMin := float64(v)
			hdp.Min = &vMin
		}
		if v, ok := pt.Max.Value(); ok {
			vMax := float64(v)
			hdp.Max = &vMax
		}
		out[i] = hdp
	}
	return out
}

func ExponentialHistogram[N int64 | float64](h metricdata.ExponentialHistogram[N]) *types.ExponentialHistogram {
	return &types.ExponentialHistogram{
		DataPoints:             ExponentialHistogramDataPoints(h.DataPoints),
		AggregationTemporality: types.AggregationTemporality(h.Temporality),
	}
}

func ExponentialHistogramDataPoints[N int64 | float64](pts []metricdata.ExponentialHistogramDataPoint[N]) []*types.ExponentialHistogramDataPoint {
	out := make([]*types.ExponentialHistogramDataPoint, len(pts))
	for i, pt := range pts {
		sum := float64(pt.Sum)
		ehdp := &types.ExponentialHistogramDataPoint{
			Attributes:        AttrIter(pt.Attributes.Iter()),
			StartTimeUnixNano: uint64(pt.StartTime.UnixNano()),
			TimeUnixNano:      uint64(pt.Time.UnixNano()),
			Count:             pt.Count,
			Sum:               &sum,
			Scale:             pt.Scale,
			Exemplars:         Exemplars(pt.Exemplars),

			Positive: ExponentialHistogramDataPointBuckets(pt.PositiveBucket),
			Negative: ExponentialHistogramDataPointBuckets(pt.NegativeBucket),
		}
		if v, ok := pt.Min.Value(); ok {
			vMin := float64(v)
			ehdp.Min = &vMin
		}
		if v, ok := pt.Max.Value(); ok {
			vMax := float64(v)
			ehdp.Max = &vMax
		}
		out[i] = ehdp
	}
	return out
}

func ExponentialHistogramDataPointBuckets(bucket metricdata.ExponentialBucket) *types.ExponentialHistogramDataPoint_Buckets {
	return &types.ExponentialHistogramDataPoint_Buckets{
		Offset:       bucket.Offset,
		BucketCounts: bucket.Counts,
	}
}

func Exemplars[N int64 | float64](exemplars []metricdata.Exemplar[N]) []*types.Exemplar {
	out := make([]*types.Exemplar, len(exemplars))
	for i, exemplar := range exemplars {
		e := &types.Exemplar{
			FilteredAttributes: KeyValues(exemplar.FilteredAttributes),
			TimeUnixNano:       uint64(exemplar.Time.UnixNano()),
			SpanId:             (*types.SpanID)(&exemplar.SpanID),
			TraceId:            (*types.TraceID)(&exemplar.TraceID),
		}
		switch v := any(exemplar.Value).(type) {
		case int64:
			e.AsInt = (*types.Exemplar_AsInt)(&v)
		case float64:
			e.AsDouble = (*types.Exemplar_AsDouble)(&v)
		}
		out[i] = e
	}
	return out
}

func Summary(s metricdata.Summary) *types.Summary {
	return &types.Summary{
		DataPoints: SummaryDataPoints(s.DataPoints),
	}
}

func SummaryDataPoints(pts []metricdata.SummaryDataPoint) []*types.SummaryDataPoint {
	out := make([]*types.SummaryDataPoint, len(pts))
	for i, pt := range pts {
		out[i] = &types.SummaryDataPoint{
			Attributes:        AttrIter(pt.Attributes.Iter()),
			StartTimeUnixNano: uint64(pt.StartTime.UnixNano()),
			TimeUnixNano:      uint64(pt.Time.UnixNano()),
			Count:             pt.Count,
			Sum:               pt.Sum,
			QuantileValues:    QuantileValues(pt.QuantileValues),
		}
	}
	return out
}

func QuantileValues(quantiles []metricdata.QuantileValue) []*types.SummaryDataPoint_ValueAtQuantile {
	out := make([]*types.SummaryDataPoint_ValueAtQuantile, len(quantiles))
	for i, q := range quantiles {
		out[i] = &types.SummaryDataPoint_ValueAtQuantile{
			Quantile: q.Quantile,
			Value:    q.Value,
		}
	}
	return out
}
