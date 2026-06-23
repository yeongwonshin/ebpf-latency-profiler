package model

import "time"

type Protocol string

const (
	ProtocolUnknown Protocol = "unknown"
	ProtocolHTTP1   Protocol = "http1"
	ProtocolGRPC    Protocol = "grpc"
)

type Direction string

const (
	DirectionIngress Direction = "ingress"
	DirectionEgress  Direction = "egress"
)

type FlowKey struct {
	PID     uint32 `json:"pid"`
	NetNS   uint64 `json:"netns"`
	Cgroup  uint64 `json:"cgroup"`
	SrcIP   string `json:"src_ip"`
	SrcPort uint16 `json:"src_port"`
	DstIP   string `json:"dst_ip"`
	DstPort uint16 `json:"dst_port"`
}

type RequestEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Flow      FlowKey   `json:"flow"`
	Direction Direction `json:"direction"`
	Protocol  Protocol  `json:"protocol"`
	Method    string    `json:"method,omitempty"`
	Path      string    `json:"path,omitempty"`
	Status    int       `json:"status,omitempty"`
	RPCSystem string    `json:"rpc_system,omitempty"`
	RPCSvc    string    `json:"rpc_service,omitempty"`
	RPCMethod string    `json:"rpc_method,omitempty"`
	Payload   []byte    `json:"-"`
}

type LatencySample struct {
	Timestamp time.Time     `json:"timestamp"`
	Source    string        `json:"source"`
	Target    string        `json:"target"`
	Operation string        `json:"operation"`
	Protocol  Protocol      `json:"protocol"`
	Status    int           `json:"status"`
	Latency   time.Duration `json:"latency"`
}

type EdgeStats struct {
	Source    string  `json:"source"`
	Target    string  `json:"target"`
	Operation string  `json:"operation"`
	Protocol  string  `json:"protocol"`
	Count     int64   `json:"count"`
	P50MS     float64 `json:"p50_ms"`
	P95MS     float64 `json:"p95_ms"`
	P99MS     float64 `json:"p99_ms"`
	ErrorRate float64 `json:"error_rate"`
}
