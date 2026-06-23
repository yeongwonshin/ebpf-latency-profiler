APP := profiler
BIN_DIR := bin
BPF_DIR := bpf
GO_FILES := $(shell find . -name '*.go' -not -path './.git/*')

.PHONY: all build test clean demo bpf generate fmt lint

all: build

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP) ./cmd/profiler

test:
	go test ./...

fmt:
	gofmt -w $(GO_FILES)

bpf:
	clang -O2 -g -target bpf -D__TARGET_ARCH_x86 -c $(BPF_DIR)/http_sock_trace.bpf.c -o $(BPF_DIR)/http_sock_trace.bpf.o

# Optional: generate Go bindings when bpf2go is available.
generate:
	go generate ./...

demo:
	cd deploy && docker compose up --build

clean:
	rm -rf $(BIN_DIR) $(BPF_DIR)/*.o
