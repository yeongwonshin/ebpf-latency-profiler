# Implementation Notes

## Why not parse everything in eBPF?

Kernel eBPF programs must satisfy verifier constraints. Bounded memory access, bounded loops, small stack size, and restricted helper usage make full HTTP/2 or protobuf parsing inside eBPF impractical. This project only emits safe, compact hints from eBPF and keeps complex parsing in userspace.

## Correlation Key

Suggested key:

```text
pid + netns + cgroup_id + src_ip + src_port + dst_ip + dst_port + direction + stream_hint
```

For HTTP/1.x without pipelining, the 4-tuple and timestamp window are usually enough. For HTTP/2/gRPC multiplexing, use uprobe-derived stream metadata when available.

## Percentile Strategy

MVP uses sorted samples in a sliding window for readability. Production extension should replace it with HDR Histogram or DDSketch-style quantile sketch.

## Security and Privacy

The profiler should not export full payloads. Only protocol prefix, route template, method, status, and latency should leave the host. Path redaction rules should be enabled by default.

