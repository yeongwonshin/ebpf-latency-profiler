package collector

import (
	"context"
	"time"

	"github.com/example/ebpf-latency-profiler/internal/model"
)

type Reader interface {
	Events(ctx context.Context) (<-chan model.RequestEvent, error)
	Close() error
}

// MockReader is useful for local development without CAP_BPF/CAP_SYS_ADMIN.
type MockReader struct {
	Interval time.Duration
}

func (r *MockReader) Events(ctx context.Context) (<-chan model.RequestEvent, error) {
	interval := r.Interval
	if interval <= 0 {
		interval = 500 * time.Millisecond
	}
	out := make(chan model.RequestEvent)
	go func() {
		defer close(out)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case now := <-ticker.C:
				flow := model.FlowKey{PID: 1234, NetNS: 1, Cgroup: 1, SrcIP: "10.0.0.10", SrcPort: 53000, DstIP: "10.0.0.20", DstPort: 8080}
				out <- model.RequestEvent{Timestamp: now, Flow: flow, Direction: model.DirectionEgress, Protocol: model.ProtocolHTTP1, Method: "GET", Path: "/orders/{id}", Payload: []byte("GET /orders/42 HTTP/1.1\r\n")}
				out <- model.RequestEvent{Timestamp: now.Add(80 * time.Millisecond), Flow: flow, Direction: model.DirectionIngress, Protocol: model.ProtocolHTTP1, Status: 200, Payload: []byte("HTTP/1.1 200 OK\r\n")}
			}
		}
	}()
	return out, nil
}

func (r *MockReader) Close() error { return nil }
