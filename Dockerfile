FROM golang:1.23-bookworm AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN go build -o /out/profiler ./cmd/profiler

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=build /out/profiler /usr/local/bin/profiler
COPY config/profiler.yaml /etc/ebpf-latency-profiler.yaml
ENTRYPOINT ["/usr/local/bin/profiler", "--config", "/etc/ebpf-latency-profiler.yaml"]
