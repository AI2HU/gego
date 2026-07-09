package etcd_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/AI2HU/gego/internal/jobs"
	etcdstore "github.com/AI2HU/gego/internal/jobs/etcd"
	"github.com/AI2HU/gego/internal/models"
)

func testStore(t *testing.T) (*etcdstore.Store, *etcdstore.Keys) {
	t.Helper()
	if os.Getenv("GEGO_ETCD_ENDPOINTS") == "" {
		t.Skip("set GEGO_ETCD_ENDPOINTS to run etcd integration tests")
	}

	cfg := jobs.LoadConfigFromEnv()
	cfg.Prefix = "/gego-test/" + uuid.New().String()

	client, err := etcdstore.NewClient(cfg)
	if err != nil {
		t.Fatalf("connect etcd: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	keys := etcdstore.NewKeys(cfg.Prefix)
	return etcdstore.NewStore(client, cfg), keys
}

func TestEnqueueClaimComplete(t *testing.T) {
	store, _ := testStore(t)
	ctx := context.Background()

	runID := uuid.New().String()
	scheduleID := uuid.New().String()
	jobID := uuid.New().String()

	run := &models.ScheduleRun{
		ID:         runID,
		ScheduleID: scheduleID,
		Trigger:    models.ScheduleRunTriggerManual,
		Status:     models.ScheduleRunStatusPending,
		TotalJobs:  1,
	}
	jobsList := []*models.ScheduleJob{{
		ID:          jobID,
		RunID:       runID,
		ScheduleID:  scheduleID,
		PromptIDs:   []string{uuid.New().String()},
		LLMID:       uuid.New().String(),
		Temperature: 0.7,
		Status:      models.ScheduleJobStatusPending,
		MaxAttempts: 3,
	}}

	if err := store.CreateRun(ctx, run, jobsList, "", 0); err != nil {
		t.Fatalf("create run: %v", err)
	}

	watchCh, err := store.WatchPendingJobs(ctx)
	if err != nil {
		t.Fatalf("watch: %v", err)
	}

	var pending *models.ScheduleJob
	select {
	case pending = <-watchCh:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for pending job")
	}

	claimed, err := store.ClaimJob(ctx, pending, "worker-1", 30*time.Second)
	if err != nil {
		t.Fatalf("claim: %v", err)
	}

	if err := store.CompleteJob(ctx, claimed.ID, []string{uuid.New().String()}); err != nil {
		t.Fatalf("complete: %v", err)
	}

	updated, err := store.GetRun(ctx, runID)
	if err != nil {
		t.Fatalf("get run: %v", err)
	}
	if updated.Status != models.ScheduleRunStatusCompleted {
		t.Fatalf("expected completed, got %s", updated.Status)
	}
}

func TestCronDedup(t *testing.T) {
	store, keys := testStore(t)
	ctx := context.Background()

	dedupKey := keys.Dedup("sched-1", "slot-1")

	run := &models.ScheduleRun{ID: uuid.New().String(), ScheduleID: "sched-1", Trigger: models.ScheduleRunTriggerCron, TotalJobs: 1}
	job := &models.ScheduleJob{ID: uuid.New().String(), RunID: run.ID, ScheduleID: "sched-1", PromptIDs: []string{"p1"}, LLMID: "l1", Temperature: 0.5, MaxAttempts: 3}

	if err := store.CreateRun(ctx, run, []*models.ScheduleJob{job}, dedupKey, time.Minute); err != nil {
		t.Fatalf("first create: %v", err)
	}

	run2 := &models.ScheduleRun{ID: uuid.New().String(), ScheduleID: "sched-1", Trigger: models.ScheduleRunTriggerCron, TotalJobs: 1}
	job2 := &models.ScheduleJob{ID: uuid.New().String(), RunID: run2.ID, ScheduleID: "sched-1", PromptIDs: []string{"p1"}, LLMID: "l1", Temperature: 0.5, MaxAttempts: 3}
	if err := store.CreateRun(ctx, run2, []*models.ScheduleJob{job2}, dedupKey, time.Minute); err == nil {
		t.Fatal("expected dedup failure")
	}
}
