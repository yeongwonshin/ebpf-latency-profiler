package otelx

import (
	"context"

	"github.com/example/ebpf-latency-profiler/internal/model"
)

type NoopExporter struct{}

func (NoopExporter) ExportEdgeStats(context.Context, []model.EdgeStats)      {}
func (NoopExporter) ExportInferredSpan(context.Context, model.LatencySample) {}
