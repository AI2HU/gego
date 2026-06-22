package etcd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"os"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/AI2HU/gego/internal/jobs"
)

func NewClient(cfg jobs.Config) (*clientv3.Client, error) {
	clientCfg := clientv3.Config{
		Endpoints:   cfg.Endpoints,
		DialTimeout: 5 * time.Second,
	}

	if cfg.TLSCert != "" && cfg.TLSKey != "" {
		tlsCfg, err := loadTLS(cfg)
		if err != nil {
			return nil, err
		}
		clientCfg.TLS = tlsCfg
	}

	return clientv3.New(clientCfg)
}

func loadTLS(cfg jobs.Config) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(cfg.TLSCert, cfg.TLSKey)
	if err != nil {
		return nil, fmt.Errorf("load etcd tls cert: %w", err)
	}

	tlsCfg := &tls.Config{Certificates: []tls.Certificate{cert}}

	if cfg.TLSCA != "" {
		caPEM, err := os.ReadFile(cfg.TLSCA)
		if err != nil {
			return nil, fmt.Errorf("read etcd ca: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(caPEM) {
			return nil, fmt.Errorf("parse etcd ca pem")
		}
		tlsCfg.RootCAs = pool
	}

	return tlsCfg, nil
}

type tokenState struct {
	Tokens     float64   `json:"tokens"`
	LastRefill time.Time `json:"last_refill"`
}

type RateLimiter struct {
	client  *clientv3.Client
	keys    *Keys
	perMin  int
	burst   float64
}

func NewRateLimiter(client *clientv3.Client, keys *Keys, requestsPerMinute int) *RateLimiter {
	if requestsPerMinute <= 0 {
		requestsPerMinute = 6
	}
	return &RateLimiter{
		client: client,
		keys:   keys,
		perMin: requestsPerMinute,
		burst:  1,
	}
}

func (r *RateLimiter) Wait(ctx context.Context, provider string) error {
	for {
		ok, wait, err := r.tryAcquire(ctx, provider)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
		if wait <= 0 {
			wait = 100 * time.Millisecond
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
	}
}

func (r *RateLimiter) tryAcquire(ctx context.Context, provider string) (bool, time.Duration, error) {
	key := r.keys.RateLimit(provider)
	now := time.Now().UTC()
	ratePerSec := float64(r.perMin) / 60.0

	for attempt := 0; attempt < 8; attempt++ {
		resp, err := r.client.Get(ctx, key)
		if err != nil {
			return false, 0, err
		}

		var state tokenState
		var modRev int64
		if len(resp.Kvs) == 0 {
			state = tokenState{Tokens: r.burst, LastRefill: now}
			modRev = 0
		} else {
			modRev = resp.Kvs[0].ModRevision
			if err := json.Unmarshal(resp.Kvs[0].Value, &state); err != nil {
				state = tokenState{Tokens: r.burst, LastRefill: now}
			}
		}

		elapsed := now.Sub(state.LastRefill).Seconds()
		if elapsed > 0 {
			state.Tokens += elapsed * ratePerSec
			if state.Tokens > r.burst {
				state.Tokens = r.burst
			}
			state.LastRefill = now
		}

		if state.Tokens < 1 {
			deficit := 1 - state.Tokens
			wait := time.Duration(deficit/ratePerSec*float64(time.Second)) + 10*time.Millisecond
			return false, wait, nil
		}

		state.Tokens -= 1
		encoded, err := json.Marshal(state)
		if err != nil {
			return false, 0, err
		}

		txn := r.client.Txn(ctx)
		if modRev == 0 {
			txn = txn.If(clientv3.Compare(clientv3.CreateRevision(key), "=", 0)).
				Then(clientv3.OpPut(key, string(encoded)))
		} else {
			txn = txn.If(clientv3.Compare(clientv3.ModRevision(key), "=", modRev)).
				Then(clientv3.OpPut(key, string(encoded)))
		}

		txnResp, err := txn.Commit()
		if err != nil {
			return false, 0, err
		}
		if txnResp.Succeeded {
			return true, 0, nil
		}
	}

	return false, 50 * time.Millisecond, nil
}
