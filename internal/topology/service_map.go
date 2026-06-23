package topology

import (
	"encoding/json"
	"io"
	"sort"
	"sync"

	"github.com/example/ebpf-latency-profiler/internal/model"
)

type ServiceResolver interface {
	Resolve(flow model.FlowKey) (source string, target string)
}

type StaticResolver struct {
	ByAddress map[string]string
}

func (r StaticResolver) Resolve(flow model.FlowKey) (string, string) {
	source := r.ByAddress[flow.SrcIP]
	if source == "" {
		source = flow.SrcIP
	}
	target := r.ByAddress[flow.DstIP]
	if target == "" {
		target = flow.DstIP
	}
	return source, target
}

type Map struct {
	mu    sync.RWMutex
	edges map[string]model.EdgeStats
}

func NewMap() *Map {
	return &Map{edges: make(map[string]model.EdgeStats)}
}

func (m *Map) Replace(stats []model.EdgeStats) {
	m.mu.Lock()
	defer m.mu.Unlock()
	next := make(map[string]model.EdgeStats, len(stats))
	for _, stat := range stats {
		next[key(stat)] = stat
	}
	m.edges = next
}

func (m *Map) Edges() []model.EdgeStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]model.EdgeStats, 0, len(m.edges))
	for _, edge := range m.edges {
		out = append(out, edge)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].P99MS == out[j].P99MS {
			return out[i].Count > out[j].Count
		}
		return out[i].P99MS > out[j].P99MS
	})
	return out
}

func (m *Map) WriteJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(struct {
		Edges []model.EdgeStats `json:"edges"`
	}{Edges: m.Edges()})
}

func key(e model.EdgeStats) string {
	return e.Source + "->" + e.Target + ":" + e.Operation + ":" + e.Protocol
}
