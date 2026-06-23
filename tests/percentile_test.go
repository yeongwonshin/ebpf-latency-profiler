package tests

import (
	"testing"
	"time"

	"github.com/example/ebpf-latency-profiler/internal/aggregator"
)

func TestComputePercentiles(t *testing.T) {
	samples := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		30 * time.Millisecond,
		40 * time.Millisecond,
		50 * time.Millisecond,
		60 * time.Millisecond,
		70 * time.Millisecond,
		80 * time.Millisecond,
		90 * time.Millisecond,
		100 * time.Millisecond,
	}
	got := aggregator.ComputePercentiles(samples)
	if got.P50 != 50*time.Millisecond {
		t.Fatalf("p50=%s", got.P50)
	}
	if got.P95 != 100*time.Millisecond {
		t.Fatalf("p95=%s", got.P95)
	}
	if got.P99 != 100*time.Millisecond {
		t.Fatalf("p99=%s", got.P99)
	}
}
