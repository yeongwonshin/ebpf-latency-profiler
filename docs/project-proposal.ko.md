# 프로젝트 제안서: eBPF 기반 HTTP/gRPC Latency Profiler

## 1. 개요

이 프로젝트는 Linux eBPF를 활용하여 HTTP/gRPC 기반 마이크로서비스 호출을 비침투적으로 추적하고, 서비스 간 지연 시간 관계를 자동으로 시각화하는 latency profiler입니다. 수집된 데이터는 OpenTelemetry 형식으로 export하여 Prometheus, Jaeger, Grafana 등 표준 관측성 도구와 연동합니다.

## 2. 배경

마이크로서비스 환경에서는 하나의 사용자 요청이 여러 내부 서비스 호출로 분산됩니다. 이때 지연의 원인이 특정 API, 특정 downstream 서비스, 네트워크, 런타임 내부 중 어디인지 찾기 어렵습니다. 기존 방식은 애플리케이션에 tracing SDK를 삽입해야 하므로 레거시 서비스, 외부 바이너리, 빠른 장애 대응 상황에서는 적용이 느릴 수 있습니다.

## 3. 핵심 질문

- 어떤 서비스 간 호출 edge가 p95/p99 지연을 유발하는가?
- 코드 수정 없이 HTTP/gRPC 요청 경로와 상태 코드를 추적할 수 있는가?
- eBPF 이벤트를 OpenTelemetry semantic model에 맞춰 표준화할 수 있는가?
- latency map을 실시간으로 갱신하고 장애 분석에 활용할 수 있는가?

## 4. 기능 요구사항

### 4.1 eBPF 이벤트 수집

- TCP send/receive 이벤트 수집
- process id, command, cgroup id, network namespace id 추출
- source/destination IP:port 추출
- payload prefix 샘플링으로 HTTP method/status line 식별
- ring buffer 또는 perf buffer를 통한 userspace 전달

### 4.2 HTTP 추적

- HTTP/1.x method: GET, POST, PUT, PATCH, DELETE 등 식별
- request path 추출
- response status code 추출
- request-response correlation key 생성
- latency 계산

### 4.3 gRPC 추적

- HTTP/2/gRPC 특성상 socket layer만으로는 제한적이므로 uprobe 확장 구조 제공
- Go gRPC runtime hook 지점 정의
- service/method, status code, deadline metadata 추출 확장

### 4.4 Latency Aggregation

- service edge별 latency histogram 유지
- p50/p95/p99 계산
- slow edge ranking
- time window별 집계 리셋

### 4.5 Service Map

- `source_service -> destination_service -> operation` edge 생성
- Kubernetes metadata와 프로세스 metadata를 이용한 service name enrichment
- JSON export 및 OTel metric label 제공

### 4.6 OpenTelemetry Export

- metric: `ebpf.http.server.duration`, `ebpf.rpc.client.duration`, `ebpf.service.edge.latency`
- trace: inferred span 생성
- log: slow request event 출력
- OTLP endpoint 설정 가능

## 5. 비기능 요구사항

- 낮은 overhead: payload 전체 복사가 아니라 prefix만 샘플링
- 안전성: eBPF verifier 통과 가능한 bounded loop와 fixed-size buffer 사용
- 확장성: protocol parser와 exporter를 인터페이스로 분리
- 운영성: config file, Docker Compose, dashboard 연결 제공

## 6. 기대 효과

- 코드 수정 없이 latency hot path 탐지
- 서비스 의존성 자동 파악
- OpenTelemetry 기반 vendor-neutral 관측성 확보
- SRE/Platform Engineering 포트폴리오로 활용 가능

## 7. 평가 지표

- HTTP 요청 추적 정확도
- p95/p99 계산 오차
- 프로파일러 CPU/메모리 overhead
- OTel Collector export 안정성
- service map edge 탐지율

