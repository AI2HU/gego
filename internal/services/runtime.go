package services

import (
	"context"
	"fmt"

	"github.com/AI2HU/gego/internal/db"
	"github.com/AI2HU/gego/internal/jobs"
	etcdstore "github.com/AI2HU/gego/internal/jobs/etcd"
	"github.com/AI2HU/gego/internal/llm"
	"github.com/AI2HU/gego/internal/models"
)

type Runtime struct {
	Config          jobs.Config
	Store           jobs.Store
	Enqueue         *RunEnqueueService
	Scheduler       *SchedulerService
	Executor        *PromptExecutor
	EtcdClient      interface{ Close() error }
}

func NewRuntime(database db.Database, llmRegistry *llm.Registry) (*Runtime, error) {
	cfg := jobs.LoadConfigFromEnv()

	client, err := etcdstore.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("connect etcd: %w", err)
	}

	store := etcdstore.NewStore(client, cfg)
	keys := etcdstore.NewKeys(cfg.Prefix)
	enqueue := NewRunEnqueueService(database, store, keys)
	scheduler := NewSchedulerService(store, enqueue)
	executor := NewPromptExecutor(database, llmRegistry, store.RateLimiter())

	return &Runtime{
		Config:     cfg,
		Store:      store,
		Enqueue:    enqueue,
		Scheduler:  scheduler,
		Executor:   executor,
		EtcdClient: client,
	}, nil
}

func (r *Runtime) Close() error {
	if r.EtcdClient != nil {
		return r.EtcdClient.Close()
	}
	return nil
}

func (r *Runtime) ListEnabledSchedules(ctx context.Context, database db.Database) ([]*models.Schedule, error) {
	enabled := true
	return database.ListSchedules(ctx, &enabled)
}

func (r *Runtime) NewWorkerService(database db.Database) *WorkerService {
	return NewWorkerService(database, r.Store, r.Executor, r.Config)
}
