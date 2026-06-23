// SPDX-License-Identifier: GPL-2.0
// gRPC uprobe extension point. This file documents the intended shape of the
// probe and is deliberately conservative because Go symbol names and ABI details
// change across versions.

#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>
#include "common.h"

char LICENSE[] SEC("license") = "GPL";

struct grpc_event {
    __u64 ts_ns;
    __u32 pid;
    __u64 cgroup_id;
    char comm[TASK_COMM_LEN];
    char full_method[128];
    __s32 status_code;
};

struct {
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 1 << 22);
} grpc_events SEC(".maps");

// Attach target examples:
// - google.golang.org/grpc.(*ClientConn).Invoke
// - google.golang.org/grpc.(*Server).processUnaryRPC
// Resolve actual symbol names with: go tool nm <binary> | grep grpc
SEC("uprobe/grpc_client_invoke")
int grpc_client_invoke(struct pt_regs *ctx) {
    struct grpc_event *event = bpf_ringbuf_reserve(&grpc_events, sizeof(*event), 0);
    if (!event) {
        return 0;
    }
    __builtin_memset(event, 0, sizeof(*event));
    event->ts_ns = bpf_ktime_get_ns();
    event->pid = (__u32)(bpf_get_current_pid_tgid() >> 32);
    event->cgroup_id = bpf_get_current_cgroup_id();
    bpf_get_current_comm(&event->comm, sizeof(event->comm));
    event->status_code = -1;
    bpf_ringbuf_submit(event, 0);
    return 0;
}
