module github.com/example/ebpf-latency-profiler

go 1.23

require (
	github.com/cilium/ebpf v0.17.3
	go.opentelemetry.io/otel v1.38.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.38.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.38.0
	go.opentelemetry.io/otel/sdk v1.38.0
	go.opentelemetry.io/otel/sdk/metric v1.38.0
	google.golang.org/grpc v1.73.0
	gopkg.in/yaml.v3 v3.0.1
)
