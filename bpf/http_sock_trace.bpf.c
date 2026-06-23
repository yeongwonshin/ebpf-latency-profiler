// SPDX-License-Identifier: GPL-2.0
// Socket/syscall layer HTTP prefix sampler.
// Build example:
//   clang -O2 -g -target bpf -D__TARGET_ARCH_x86 -c bpf/http_sock_trace.bpf.c -o bpf/http_sock_trace.bpf.o

#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>
#include <bpf/bpf_core_read.h>
#include "common.h"

char LICENSE[] SEC("license") = "GPL";

struct {
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 1 << 24);
} events SEC(".maps");

static __always_inline __u8 guess_protocol(const char *buf, __u32 len) {
    if (len < 4) {
        return PROTO_UNKNOWN;
    }
    if ((buf[0] == 'G' && buf[1] == 'E' && buf[2] == 'T' && buf[3] == ' ') ||
        (buf[0] == 'P' && buf[1] == 'O' && buf[2] == 'S' && buf[3] == 'T') ||
        (buf[0] == 'P' && buf[1] == 'U' && buf[2] == 'T' && buf[3] == ' ') ||
        (buf[0] == 'H' && buf[1] == 'T' && buf[2] == 'T' && buf[3] == 'P')) {
        return PROTO_HTTP1;
    }
    return PROTO_UNKNOWN;
}

// Minimal tracepoint signature for sendto payload prefix capture. Production code
// should enrich the flow_id from socket metadata using sockfd lookup helpers or
// kprobes around tcp_sendmsg/tcp_recvmsg.
SEC("tracepoint/syscalls/sys_enter_sendto")
int trace_sendto(struct trace_event_raw_sys_enter *ctx) {
    const void *user_buf = (const void *)ctx->args[1];
    __u64 size = ctx->args[2];
    if (user_buf == 0 || size == 0) {
        return 0;
    }

    struct http_event *event = bpf_ringbuf_reserve(&events, sizeof(*event), 0);
    if (!event) {
        return 0;
    }

    __builtin_memset(event, 0, sizeof(*event));
    event->ts_ns = bpf_ktime_get_ns();
    event->flow.pid = (__u32)(bpf_get_current_pid_tgid() >> 32);
    event->flow.cgroup_id = bpf_get_current_cgroup_id();
    event->direction = DIR_EGRESS;
    event->payload_len = size < PREFIX_LEN ? size : PREFIX_LEN;
    bpf_get_current_comm(&event->comm, sizeof(event->comm));
    bpf_probe_read_user(event->payload, event->payload_len, user_buf);
    event->protocol_hint = guess_protocol(event->payload, event->payload_len);

    bpf_ringbuf_submit(event, 0);
    return 0;
}

SEC("tracepoint/syscalls/sys_enter_recvfrom")
int trace_recvfrom(struct trace_event_raw_sys_enter *ctx) {
    // recvfrom entry does not yet have payload data. Use sys_exit_recvfrom with
    // a temporary map keyed by pid/tid in a production implementation.
    return 0;
}
