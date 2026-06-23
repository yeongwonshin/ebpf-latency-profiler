# eBPF 기반 HTTP/gRPC Latency Profiler

> 애플리케이션 코드 수정 없이 Linux eBPF probe로 HTTP/gRPC 요청 흐름을 관찰하고, service-to-service latency map과 p50/p95/p99 지표를 OpenTelemetry로 내보내는 관측성 프로젝트입니다.

## 1. 프로젝트 목표

마이크로서비스 환경에서 “어느 서비스 호출이 느린가?”, “p95/p99 지연이 어디서 튀는가?”, “클라이언트/서버 어느 쪽에서 병목이 발생하는가?”를 코드 삽입 없이 파악하는 것이 목표입니다.

핵심 구현 범위는 다음과 같습니다.

- **eBPF 기반 요청 추적**
  - 1차: socket/syscall layer에서 TCP flow와 payload prefix를 관찰합니다.
  - 2차: uprobe 기반으로 Go `net/http`, `google.golang.org/grpc` 런타임 함수 hook을 추가합니다.
- **HTTP/gRPC latency 계산**
  - request/response timestamp를 correlation key로 매칭합니다.
  - HTTP method/path/status, gRPC service/method/status를 추출합니다.
- **service-to-service latency map 생성**
  - `src_service -> dst_service -> route/rpc` edge를 만들고 latency percentile을 붙입니다.
- **p50/p95/p99 집계**
  - sliding window 기반 histogram/quantile estimator를 사용합니다.
- **OpenTelemetry 연동**
  - OTLP exporter로 metrics/traces/logs를 OTel Collector에 전송합니다.
  - Prometheus, Jaeger, Grafana와 연결 가능한 구성을 제공합니다.

## 2. 발전된 프로젝트 주제

### 제목

**Zero-Instrumentation eBPF HTTP/gRPC Latency Profiler with OpenTelemetry Service Map**

### 문제 정의

기존 APM은 SDK 삽입, 코드 수정, 재배포가 필요한 경우가 많습니다. 이 프로젝트는 eBPF를 이용해 커널 또는 런타임 boundary에서 요청을 관찰하여 기존 서비스 코드 변경 없이 latency, dependency graph, tail latency 이상 징후를 수집합니다.

### 차별화 포인트

1. **비침투형 관측성**: 애플리케이션 코드를 수정하지 않고 요청 흐름을 관찰합니다.
2. **계층형 probe 전략**: socket layer는 범용성이 높고, uprobe는 프로토콜 해석 정확도를 높입니다.
3. **서비스 맵 자동 생성**: IP:port, Kubernetes metadata, 프로세스 정보를 조합해 서비스 간 호출 그래프를 만듭니다.
4. **Tail latency 중심 분석**: 평균이 아니라 p95/p99와 slow edge를 우선 노출합니다.
5. **OTel-native 설계**: vendor lock-in 없이 OTLP로 내보내고 Grafana/Jaeger/Prometheus와 연결합니다.

## 3. 아키텍처

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
                                                        +-----------------------+
                                                        | OpenTelemetry Collector|
                                                        +-----+----------+------+
                                                              |          |
                                                       Prometheus      Jaeger
                                                              |
                                                           Grafana
```

## 4. 디렉토리 구조

```text
ebpf-latency-profiler/
├── bpf/                     # eBPF C programs and shared headers
├── cmd/
│   ├── profiler/            # main agent binary
│   ├── demo-http/           # HTTP demo service
│   └── demo-grpc/           # gRPC demo service skeleton
├── config/                  # profiler and OTel Collector configs
├── deploy/                  # Docker Compose stack
├── docs/                    # Korean proposal and architecture docs
├── examples/                # sample service map output
├── internal/                # Go packages
│   ├── aggregator/          # percentile and sliding-window latency logic
│   ├── collector/           # eBPF event reader abstraction
│   ├── config/              # YAML config loader
│   ├── model/               # event/edge models
│   ├── otel/                # OpenTelemetry exporter
│   ├── protocol/            # HTTP/gRPC parser helpers
│   └── topology/            # service dependency graph
├── scripts/                 # helper scripts
├── tests/                   # unit tests and test data
├── Dockerfile
├── Makefile
└── go.mod
```

## 5. 실행 방법

### 로컬 개발 준비

```bash
./scripts/setup-dev.sh
```

### 데모 스택 실행

```bash
make demo
```

### Profiler 실행 예시

```bash
sudo ./bin/profiler --config ./config/profiler.yaml
```

### OTel Collector 포함 실행

```bash
cd deploy
sudo docker compose up --build
```

## 6. 산출물

- eBPF probe 기반 요청 이벤트 수집기
- HTTP/gRPC latency percentile exporter
- service-to-service topology JSON
- OpenTelemetry OTLP metrics/traces exporter
- Grafana/Jaeger/Prometheus 연동 가능한 demo stack
- 포트폴리오용 설계 문서와 발표용 아키텍처 설명

## 7. 구현 로드맵

| 단계 | 목표 | 산출물 |
|---|---|---|
| 1 | socket/syscall trace MVP | TCP flow event, timestamp, pid, ip:port 수집 |
| 2 | HTTP parser | method/path/status 기반 request-response 매칭 |
| 3 | latency aggregator | p50/p95/p99 sliding window 구현 |
| 4 | service map | service edge graph JSON/API 출력 |
| 5 | OpenTelemetry | OTLP metrics/traces exporter 구현 |
| 6 | gRPC uprobe | service/method/status 추출 정확도 개선 |
| 7 | dashboard | Grafana + Jaeger demo dashboard |

## 8. 한계와 확장 과제

- TLS 암호화 트래픽은 socket payload만으로 HTTP path를 볼 수 없습니다. 이 경우 uprobe, service mesh sidecar, TLS library hook, 또는 application metadata와 조합해야 합니다.
- gRPC는 HTTP/2 framing과 protobuf payload 때문에 socket layer만으로 완전 해석이 어렵습니다. 본 프로젝트는 gRPC method/status 추출을 위해 uprobe 경로를 별도 확장 지점으로 둡니다.
- production 환경에서는 eBPF verifier 제한, 커널 버전, container namespace, 권한 모델, overhead budget을 검증해야 합니다.

