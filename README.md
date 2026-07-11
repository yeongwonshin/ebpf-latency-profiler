# eBPF-based HTTP/gRPC Latency Profiler

> An observability project that monitors HTTP and gRPC request flows using Linux eBPF probes without modifying application code, then exports service-to-service latency maps and p50/p95/p99 metrics through OpenTelemetry.

## 1. Project Goals

The goal is to answer questions such as “Which service call is slow?”, “Where are p95 and p99 latencies spiking?”, and “Is the bottleneck on the client side or the server side?” in microservice environments without adding application instrumentation.

The core implementation scope includes the following:

- **eBPF-based Request Tracing**
  - Phase 1: Observe TCP flows and payload prefixes at the socket and syscall layers.
  - Phase 2: Add uprobe-based hooks for Go `net/http` and `google.golang.org/grpc` runtime functions.
- **HTTP/gRPC Latency Calculation**
  - Match request and response timestamps using correlation keys.
  - Extract HTTP method, path, and status, as well as gRPC service, method, and status.
- **Service-to-Service Latency Map**
  - Build `src_service -> dst_service -> route/rpc` edges and attach latency percentiles.
- **p50/p95/p99 Aggregation**
  - Use sliding-window histograms or quantile estimators.
- **OpenTelemetry Integration**
  - Export metrics, traces, and logs to an OpenTelemetry Collector through OTLP.
  - Provide configurations compatible with Prometheus, Jaeger, and Grafana.

## 2. Refined Project Concept

### Title

**Zero-Instrumentation eBPF HTTP/gRPC Latency Profiler with OpenTelemetry Service Map**

### Problem Statement

Traditional APM solutions often require SDK integration, source-code changes, and redeployment. This project uses eBPF to observe requests at kernel or runtime boundaries, enabling collection of latency, dependency graphs, and tail-latency anomalies without modifying existing service code.

### Key Differentiators

1. **Non-intrusive Observability**: Observe request flows without changing application code.
2. **Layered Probe Strategy**: Socket-layer probes provide broad coverage, while uprobes improve protocol-level accuracy.
3. **Automatic Service Map Generation**: Combine IP:port data, Kubernetes metadata, and process information to build service dependency graphs.
4. **Tail-latency-first Analysis**: Prioritize p95, p99, and slow edges rather than average latency.
5. **OTel-native Design**: Export through OTLP without vendor lock-in and integrate with Grafana, Jaeger, and Prometheus.

## 3. Architecture

```text
+-------------------+        ringbuf/perfbuf        +-----------------------+
| eBPF Programs     | --------------------------->  | Go Profiler Agent     |
| - socket trace    |                               | - event decoder       |
| - syscall trace   |                               | - protocol parser     |
| - uprobes         |                               | - latency aggregator  |
+-------------------+                               | - service topology    |
                                                    | - OTel exporter       |
                                                    +----------+------------+
                                                               |
                                                               | OTLP gRPC/HTTP
                                                               v
                                                    +------------------------+
                                                    | OpenTelemetry Collector|
                                                    +-----+----------+-------+
                                                          |          |
                                                   Prometheus      Jaeger
                                                          |
                                                       Grafana
```

## 4. Directory Structure

```text
ebpf-latency-profiler/
├── bpf/                     # eBPF C programs and shared headers
├── cmd/
│   ├── profiler/            # Main agent binary
│   ├── demo-http/           # HTTP demo service
│   └── demo-grpc/           # gRPC demo service scaffold
├── config/                  # Profiler and OTel Collector configurations
├── deploy/                  # Docker Compose stack
├── docs/                    # Proposal and architecture documentation
├── examples/                # Sample service map output
├── internal/                # Go packages
│   ├── aggregator/          # Percentile and sliding-window latency logic
│   ├── collector/           # eBPF event reader abstraction
│   ├── config/              # YAML configuration loader
│   ├── model/               # Event and edge models
│   ├── otel/                # OpenTelemetry exporter
│   ├── protocol/            # HTTP/gRPC parser helpers
│   └── topology/            # Service dependency graph
├── scripts/                 # Helper scripts
├── tests/                   # Unit tests and test data
├── Dockerfile
├── Makefile
└── go.mod
```

## 5. Running the Project

### Prepare the Local Development Environment

```bash
./scripts/setup-dev.sh
```

### Run the Demo Stack

```bash
make demo
```

### Example Profiler Command

```bash
sudo ./bin/profiler --config ./config/profiler.yaml
```

### Run with the OpenTelemetry Collector

```bash
cd deploy
sudo docker compose up --build
```

## 6. Deliverables

- Request event collector based on eBPF probes
- HTTP/gRPC latency percentile exporter
- Service-to-service topology JSON
- OpenTelemetry OTLP metrics and traces exporter
- Demo stack compatible with Grafana, Jaeger, and Prometheus
- Portfolio-ready design documents and presentation-oriented architecture materials

## 7. Implementation Roadmap

| Phase | Goal | Deliverable |
|---|---|---|
| 1 | Socket/syscall tracing MVP | Collect TCP flow events, timestamps, PIDs, and IP:port pairs |
| 2 | HTTP parser | Match requests and responses by method, path, and status |
| 3 | Latency aggregator | Implement p50/p95/p99 sliding-window aggregation |
| 4 | Service map | Export a service-edge graph through JSON or an API |
| 5 | OpenTelemetry | Implement OTLP metrics and traces exporters |
| 6 | gRPC uprobe | Improve extraction accuracy for service, method, and status |
| 7 | Dashboard | Provide a Grafana and Jaeger demo dashboard |

## 8. Limitations and Future Work

- TLS-encrypted traffic prevents HTTP paths from being observed through socket payloads alone. In such cases, uprobes, service-mesh sidecars, TLS library hooks, or application metadata must be combined.
- gRPC is difficult to decode completely at the socket layer because of HTTP/2 framing and protobuf payloads. This project therefore treats uprobes as a dedicated extension path for extracting gRPC method and status information.
- Production environments require validation of eBPF verifier constraints, kernel versions, container namespaces, permission models, and overhead budgets.
