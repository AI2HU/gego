package jobs

import (
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultPrefix            = "/gego"
	defaultWorkerConcurrency = 8
	defaultJobLease          = 120 * time.Second
	defaultMaxRetries        = 3
	defaultReclaimInterval   = 30 * time.Second
	defaultRunRetention      = 720 * time.Hour
	defaultCompactInterval   = time.Hour
	defaultRequestsPerMinute = 6
)

type Config struct {
	Endpoints          []string
	Prefix             string
	TLSCert            string
	TLSKey             string
	TLSCA              string
	WorkerConcurrency  int
	JobLease           time.Duration
	MaxRetries         int
	ReclaimInterval    time.Duration
	RunRetention       time.Duration
	CompactInterval    time.Duration
	RequestsPerMinute  int
}

func LoadConfigFromEnv() Config {
	cfg := Config{
		Endpoints:         []string{"127.0.0.1:2379"},
		Prefix:            defaultPrefix,
		WorkerConcurrency: defaultWorkerConcurrency,
		JobLease:          defaultJobLease,
		MaxRetries:        defaultMaxRetries,
		ReclaimInterval:   defaultReclaimInterval,
		RunRetention:      defaultRunRetention,
		CompactInterval:   defaultCompactInterval,
		RequestsPerMinute: defaultRequestsPerMinute,
	}

	if endpoints := strings.TrimSpace(os.Getenv("GEGO_ETCD_ENDPOINTS")); endpoints != "" {
		parts := strings.Split(endpoints, ",")
		cfg.Endpoints = make([]string, 0, len(parts))
		for _, part := range parts {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				cfg.Endpoints = append(cfg.Endpoints, trimmed)
			}
		}
	}

	if prefix := strings.TrimSpace(os.Getenv("GEGO_ETCD_PREFIX")); prefix != "" {
		cfg.Prefix = prefix
	}

	cfg.TLSCert = os.Getenv("GEGO_ETCD_TLS_CERT")
	cfg.TLSKey = os.Getenv("GEGO_ETCD_TLS_KEY")
	cfg.TLSCA = os.Getenv("GEGO_ETCD_TLS_CA")

	if v := os.Getenv("GEGO_WORKER_CONCURRENCY"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.WorkerConcurrency = n
		}
	}

	if v := os.Getenv("GEGO_WORKER_JOB_LEASE"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			cfg.JobLease = d
		}
	}

	if v := os.Getenv("GEGO_JOB_MAX_RETRIES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.MaxRetries = n
		}
	}

	if v := os.Getenv("GEGO_RECLAIM_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			cfg.ReclaimInterval = d
		}
	}

	if v := os.Getenv("GEGO_RUN_RETENTION"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			cfg.RunRetention = d
		}
	}

	if v := os.Getenv("GEGO_RUN_COMPACT_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			cfg.CompactInterval = d
		}
	}

	return cfg
}
