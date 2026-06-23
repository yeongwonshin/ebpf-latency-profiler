package tests

import (
	"testing"

	"github.com/example/ebpf-latency-profiler/internal/protocol"
)

func TestParseHTTP1Request(t *testing.T) {
	hint := protocol.ParseHTTP1Prefix([]byte("GET /orders/42?debug=true HTTP/1.1\r\nHost: example\r\n"))
	if !hint.IsRequest || hint.Method != "GET" || hint.Path != "/orders/42" {
		t.Fatalf("unexpected hint: %+v", hint)
	}
}

func TestParseHTTP1Response(t *testing.T) {
	hint := protocol.ParseHTTP1Prefix([]byte("HTTP/1.1 503 Service Unavailable\r\n"))
	if !hint.IsResponse || hint.Status != 503 {
		t.Fatalf("unexpected hint: %+v", hint)
	}
}

func TestParseGRPCMethod(t *testing.T) {
	hint := protocol.ParseGRPCMethod("/shop.orders.v1.OrderService/GetOrder")
	if hint.Service != "shop.orders.v1.OrderService" || hint.Method != "GetOrder" {
		t.Fatalf("unexpected hint: %+v", hint)
	}
}
