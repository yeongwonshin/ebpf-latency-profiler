#!/usr/bin/env bash
set -euo pipefail

cat <<'MSG'
Install these packages on a Linux development host:

Debian/Ubuntu:
  sudo apt-get update
  sudo apt-get install -y clang llvm make gcc libbpf-dev linux-headers-$(uname -r) bpftool

Fedora:
  sudo dnf install -y clang llvm make gcc libbpf-devel kernel-devel bpftool

Then run:
  go mod download
  make test
  make build

For eBPF loading, run the profiler with sufficient privileges, for example CAP_BPF/CAP_PERFMON or sudo depending on your kernel and distribution.
MSG
