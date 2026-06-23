#ifndef __EBPF_LATENCY_PROFILER_COMMON_H
#define __EBPF_LATENCY_PROFILER_COMMON_H

#define TASK_COMM_LEN 16
#define PREFIX_LEN 160

#define DIR_INGRESS 1
#define DIR_EGRESS  2

#define PROTO_UNKNOWN 0
#define PROTO_HTTP1   1
#define PROTO_GRPC    2

struct flow_id {
    __u32 pid;
    __u64 netns;
    __u64 cgroup_id;
    __u32 src_ip4;
    __u32 dst_ip4;
    __u16 src_port;
    __u16 dst_port;
};

struct http_event {
    __u64 ts_ns;
    struct flow_id flow;
    __u8 direction;
    __u8 protocol_hint;
    __u16 payload_len;
    char comm[TASK_COMM_LEN];
    char payload[PREFIX_LEN];
};

#endif
