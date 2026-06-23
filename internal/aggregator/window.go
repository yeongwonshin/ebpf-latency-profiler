package aggregator

import (
	"sync"
	"time"

	"github.com/example/ebpf-latency-profiler/internal/model"
)

type edgeKey struct {
	source    string
	target    string
	operation string
	protocol  model.Protocol
}

type storedSample struct {
	at      time.Time
	latency time.Duration
	status  int
}

type SlidingWindow struct {
	mu      sync.RWMutex
	window  time.Duration
	samples map[edgeKey][]storedSample
}

func NewSlidingWindow(window time.Duration) *SlidingWindow {
	return &SlidingWindow{
		window:  window,
		samples: make(map[edgeKey][]storedSample),
	}
}

func (w *SlidingWindow) Add(sample model.LatencySample) {
	w.mu.Lock()
	defer w.mu.Unlock()
	key := edgeKey{
		source:    sample.Source,
		target:    sample.Target,
		operation: sample.Operation,
		protocol:  sample.Protocol,
	}
	w.samples[key] = append(w.samples[key], storedSample{
		at:      sample.Timestamp,
		latency: sample.Latency,
		status:  sample.Status,
	})
	w.compactLocked(time.Now())
}

func (w *SlidingWindow) Snapshot(now time.Time) []model.EdgeStats {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.compactLocked(now)
	out := make([]model.EdgeStats, 0, len(w.samples))
	for key, values := range w.samples {
		latencies := make([]time.Duration, 0, len(values))
		errors := 0
		for _, value := range values {
			latencies = append(latencies, value.latency)
			if value.status >= 500 || value.status < 0 {
				errors++
			}
		}
		p := ComputePercentiles(latencies)
		errorRate := 0.0
		if len(values) > 0 {
			errorRate = float64(errors) / float64(len(values))
		}
		out = append(out, model.EdgeStats{
			Source:    key.source,
			Target:    key.target,
			Operation: key.operation,
			Protocol:  string(key.protocol),
			Count:     int64(len(values)),
			P50MS:     durationMS(p.P50),
			P95MS:     durationMS(p.P95),
			P99MS:     durationMS(p.P99),
			ErrorRate: errorRate,
		})
	}
	return out
}

func (w *SlidingWindow) compactLocked(now time.Time) {
	cutoff := now.Add(-w.window)
	for key, values := range w.samples {
		kept := values[:0]
		for _, value := range values {
			if value.at.After(cutoff) || value.at.Equal(cutoff) {
				kept = append(kept, value)
			}
		}
		if len(kept) == 0 {
			delete(w.samples, key)
			continue
		}
		w.samples[key] = kept
	}
}

func durationMS(d time.Duration) float64 {
	return float64(d.Microseconds()) / 1000.0
}
