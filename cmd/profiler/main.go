package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/example/ebpf-latency-profiler/internal/aggregator"
	"github.com/example/ebpf-latency-profiler/internal/collector"
	"github.com/example/ebpf-latency-profiler/internal/config"
	"github.com/example/ebpf-latency-profiler/internal/model"
	otelx "github.com/example/ebpf-latency-profiler/internal/otel"
	"github.com/example/ebpf-latency-profiler/internal/topology"
)

type metricExporter interface {
	ExportEdgeStats(context.Context, []model.EdgeStats)
	ExportInferredSpan(context.Context, model.LatencySample)
}

type pendingRequest struct {
	at        time.Time
	method    string
	operation string
	flow      model.FlowKey
	protocol  model.Protocol
}

func main() {
	configPath := flag.String("config", "", "path to profiler config")
	mock := flag.Bool("mock", true, "use mock events instead of loading eBPF programs")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	window := aggregator.NewSlidingWindow(cfg.Agent.Window)
	svcMap := topology.NewMap()
	resolver := topology.StaticResolver{ByAddress: map[string]string{
		"10.0.0.10": "frontend",
		"10.0.0.20": "orders",
	}}

	var reader collector.Reader
	if *mock {
		reader = &collector.MockReader{Interval: 750 * time.Millisecond}
	} else {
		reader = &collector.RingbufReader{ObjectPath: "./bpf/http_sock_trace.bpf.o"}
	}
	defer reader.Close()

	events, err := reader.Events(ctx)
	if err != nil {
		slog.Error("start collector", "error", err)
		os.Exit(1)
	}

	var exporter metricExporter = otelx.NoopExporter{}
	if cfg.OpenTelemetry.Endpoint != "" {
		otelExporter, err := otelx.New(ctx, otelx.Options{Endpoint: cfg.OpenTelemetry.Endpoint, Insecure: cfg.OpenTelemetry.Insecure, ServiceName: cfg.Agent.ServiceName})
		if err != nil {
			slog.Warn("otel exporter disabled", "error", err)
		} else {
			exporter = otelExporter
			defer otelExporter.Shutdown(context.Background())
		}
	}

	pending := make(map[string]pendingRequest)
	flushTicker := time.NewTicker(cfg.Agent.FlushInterval)
	defer flushTicker.Stop()
	expireTicker := time.NewTicker(30 * time.Second)
	defer expireTicker.Stop()

	slog.Info("profiler started", "mock", *mock, "window", cfg.Agent.Window, "flush_interval", cfg.Agent.FlushInterval)

	for {
		select {
		case <-ctx.Done():
			flush(ctx, window, svcMap, exporter)
			return
		case ev, ok := <-events:
			if !ok {
				flush(ctx, window, svcMap, exporter)
				return
			}
			handleEvent(ctx, ev, pending, resolver, window, exporter)
		case <-flushTicker.C:
			flush(ctx, window, svcMap, exporter)
		case now := <-expireTicker.C:
			expirePending(pending, now, 15*time.Second)
		}
	}
}

func handleEvent(ctx context.Context, ev model.RequestEvent, pending map[string]pendingRequest, resolver topology.ServiceResolver, window *aggregator.SlidingWindow, exporter metricExporter) {
	key := flowID(ev.Flow)
	if ev.Method != "" || ev.Path != "" || ev.RPCMethod != "" {
		operation := ev.Path
		if operation == "" && ev.RPCMethod != "" {
			operation = ev.RPCSvc + "/" + ev.RPCMethod
		}
		if operation == "" {
			operation = ev.Method
		}
		pending[key] = pendingRequest{at: ev.Timestamp, method: ev.Method, operation: operation, flow: ev.Flow, protocol: ev.Protocol}
		return
	}
	if ev.Status == 0 {
		return
	}
	req, ok := pending[key]
	if !ok {
		return
	}
	delete(pending, key)
	src, dst := resolver.Resolve(req.flow)
	sample := model.LatencySample{
		Timestamp: ev.Timestamp,
		Source:    src,
		Target:    dst,
		Operation: req.operation,
		Protocol:  req.protocol,
		Status:    ev.Status,
		Latency:   ev.Timestamp.Sub(req.at),
	}
	window.Add(sample)
	exporter.ExportInferredSpan(ctx, sample)
}

func flush(ctx context.Context, window *aggregator.SlidingWindow, svcMap *topology.Map, exporter metricExporter) {
	stats := window.Snapshot(time.Now())
	svcMap.Replace(stats)
	exporter.ExportEdgeStats(ctx, stats)
	if len(stats) == 0 {
		return
	}
	body, _ := json.MarshalIndent(struct {
		Edges []model.EdgeStats `json:"edges"`
	}{Edges: svcMap.Edges()}, "", "  ")
	fmt.Println(string(body))
}

func expirePending(pending map[string]pendingRequest, now time.Time, ttl time.Duration) {
	for key, req := range pending {
		if now.Sub(req.at) > ttl {
			delete(pending, key)
		}
	}
}

func flowID(f model.FlowKey) string {
	return fmt.Sprintf("%d:%d:%d:%s:%d:%s:%d", f.PID, f.NetNS, f.Cgroup, f.SrcIP, f.SrcPort, f.DstIP, f.DstPort)
}
