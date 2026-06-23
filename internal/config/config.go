package config

import (
	"errors"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Agent struct {
		ServiceName       string        `yaml:"service_name"`
		Window            time.Duration `yaml:"window"`
		FlushInterval     time.Duration `yaml:"flush_interval"`
		PayloadPrefixSize int           `yaml:"payload_prefix_size"`
	} `yaml:"agent"`
	EBPF struct {
		Mode          string   `yaml:"mode"`
		EnableUprobes bool     `yaml:"enable_uprobes"`
		TargetPIDs    []uint32 `yaml:"target_pids"`
	} `yaml:"ebpf"`
	OpenTelemetry struct {
		Endpoint string `yaml:"endpoint"`
		Insecure bool   `yaml:"insecure"`
	} `yaml:"opentelemetry"`
	Redaction struct {
		Enabled      bool     `yaml:"enabled"`
		PathPatterns []string `yaml:"path_patterns"`
	} `yaml:"redaction"`
}

func Load(path string) (Config, error) {
	cfg := Default()
	if path == "" {
		return cfg, nil
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	if err := yaml.Unmarshal(body, &cfg); err != nil {
		return Config{}, err
	}
	if cfg.Agent.Window <= 0 {
		return Config{}, errors.New("agent.window must be positive")
	}
	if cfg.Agent.FlushInterval <= 0 {
		return Config{}, errors.New("agent.flush_interval must be positive")
	}
	return cfg, nil
}

func Default() Config {
	var cfg Config
	cfg.Agent.ServiceName = "ebpf-latency-profiler"
	cfg.Agent.Window = 5 * time.Minute
	cfg.Agent.FlushInterval = 10 * time.Second
	cfg.Agent.PayloadPrefixSize = 160
	cfg.EBPF.Mode = "socket"
	cfg.EBPF.EnableUprobes = false
	cfg.OpenTelemetry.Endpoint = "localhost:4317"
	cfg.OpenTelemetry.Insecure = true
	cfg.Redaction.Enabled = true
	cfg.Redaction.PathPatterns = []string{"/users/{id}", "/orders/{id}"}
	return cfg
}
