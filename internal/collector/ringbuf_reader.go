package collector

import (
	"context"
	"errors"

	"github.com/example/ebpf-latency-profiler/internal/model"
)

// RingbufReader is the production extension point. In a complete build, this
// loads the compiled BPF object, attaches probes, and decodes records from a
// BPF ring buffer. It is intentionally isolated from the rest of the agent so
// the repository can be developed and tested without requiring kernel privileges.
type RingbufReader struct {
	ObjectPath string
}

func (r *RingbufReader) Events(ctx context.Context) (<-chan model.RequestEvent, error) {
	return nil, errors.New("ringbuf reader is a scaffold: generate BPF bindings with bpf2go and implement decoder")
}

func (r *RingbufReader) Close() error { return nil }
