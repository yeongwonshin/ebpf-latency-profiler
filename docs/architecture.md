# Architecture

## Data Plane

The data plane consists of eBPF programs attached to socket/syscall tracepoints and optional uprobes. The eBPF side emits compact events instead of doing heavy parsing in kernel space.

Recommended event design:

- flow identity: pid, netns, cgroup id, src/dst address, src/dst port
- timing: monotonic timestamp in nanoseconds
- direction: ingress or egress
- protocol hint: unknown, http1_request, http1_response, grpc_hint
- payload prefix: fixed-size prefix for userspace parsing

## Control Plane

The Go agent performs the expensive work:

1. reads ring buffer events
2. parses protocol prefix
3. enriches events with service metadata
4. correlates request and response
5. updates latency histograms
6. exports metrics/traces/logs via OTLP

## Probe Strategy

### Socket/syscall layer

Pros:

- language agnostic
- works without app source code
- useful for legacy binaries

Cons:

- TLS payload is encrypted
- HTTP/2 parsing is limited
- request/response correlation is probabilistic under multiplexing

### Uprobe layer

Pros:

- can observe higher-level runtime data
- better for TLS and gRPC metadata
- better request/response correlation

Cons:

- language/runtime specific
- symbol/version compatibility issues
- higher maintenance cost

## OpenTelemetry Mapping

### Metrics

- `ebpf.http.client.duration`: HTTP client latency histogram
- `ebpf.http.server.duration`: HTTP server latency histogram
- `ebpf.rpc.client.duration`: RPC/gRPC client latency histogram
- `ebpf.service.edge.duration`: service-to-service edge latency histogram

### Trace Attributes

- `service.name`
- `http.request.method`
- `url.path`
- `http.response.status_code`
- `rpc.system=grpc`
- `rpc.service`
- `rpc.method`
- `network.peer.address`
- `network.peer.port`

## Failure Modes

- Missing response event: expire pending request by timeout.
- Duplicate event: use flow key + timestamp window de-duplication.
- Unknown service name: fall back to `pid:comm` or `ip:port`.
- TLS traffic: mark route as encrypted and rely on uprobes for path-level metadata.

