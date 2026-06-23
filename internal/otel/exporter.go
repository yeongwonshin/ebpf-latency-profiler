package otelx

import (
	"context"
	"log/slog"
	"time"

	"github.com/example/ebpf-latency-profiler/internal/model"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
	"go.opentelemetry.io/otel/trace"
)

type Exporter struct {
	meterProvider *sdkmetric.MeterProvider
	traceProvider *sdktrace.TracerProvider
	histogram     metric.Float64Histogram
	tracer        trace.Tracer
}

type Options struct {
	Endpoint    string
	Insecure    bool
	ServiceName string
}

func New(ctx context.Context, opts Options) (*Exporter, error) {
	metricOpts := []otlpmetricgrpc.Option{otlpmetricgrpc.WithEndpoint(opts.Endpoint)}
	traceOpts := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(opts.Endpoint)}
	if opts.Insecure {
		metricOpts = append(metricOpts, otlpmetricgrpc.WithInsecure())
		traceOpts = append(traceOpts, otlptracegrpc.WithInsecure())
	}
	metricExporter, err := otlpmetricgrpc.New(ctx, metricOpts...)
	if err != nil {
		return nil, err
	}
	traceExporter, err := otlptracegrpc.New(ctx, traceOpts...)
	if err != nil {
		return nil, err
	}
	res, err := resource.New(ctx, resource.WithAttributes(semconv.ServiceName(opts.ServiceName)))
	if err != nil {
		return nil, err
	}
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter, sdkmetric.WithInterval(10*time.Second))),
	)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(traceExporter),
	)
	otel.SetMeterProvider(mp)
	otel.SetTracerProvider(tp)
	meter := mp.Meter("ebpf-latency-profiler")
	hist, err := meter.Float64Histogram(
		"ebpf.service.edge.duration",
		metric.WithDescription("Observed service-to-service request latency from eBPF events"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return nil, err
	}
	return &Exporter{meterProvider: mp, traceProvider: tp, histogram: hist, tracer: tp.Tracer("ebpf-latency-profiler")}, nil
}

func (e *Exporter) ExportEdgeStats(ctx context.Context, stats []model.EdgeStats) {
	for _, stat := range stats {
		attrs := metric.WithAttributes(
			attribute.String("source.service", stat.Source),
			attribute.String("target.service", stat.Target),
			attribute.String("operation", stat.Operation),
			attribute.String("network.protocol.name", stat.Protocol),
		)
		e.histogram.Record(ctx, stat.P50MS, attrs)
		e.histogram.Record(ctx, stat.P95MS, attrs)
		e.histogram.Record(ctx, stat.P99MS, attrs)
		if stat.P99MS > 1000 {
			slog.Warn("slow service edge detected", "source", stat.Source, "target", stat.Target, "operation", stat.Operation, "p99_ms", stat.P99MS)
		}
	}
}

func (e *Exporter) ExportInferredSpan(ctx context.Context, sample model.LatencySample) {
	_, span := e.tracer.Start(ctx, sample.Operation,
		trace.WithTimestamp(sample.Timestamp.Add(-sample.Latency)),
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("source.service", sample.Source),
			attribute.String("target.service", sample.Target),
			attribute.String("network.protocol.name", string(sample.Protocol)),
			attribute.Int("http.response.status_code", sample.Status),
		),
	)
	span.End(trace.WithTimestamp(sample.Timestamp))
}

func (e *Exporter) Shutdown(ctx context.Context) error {
	if e == nil {
		return nil
	}
	if e.meterProvider != nil {
		if err := e.meterProvider.Shutdown(ctx); err != nil {
			return err
		}
	}
	if e.traceProvider != nil {
		return e.traceProvider.Shutdown(ctx)
	}
	return nil
}
