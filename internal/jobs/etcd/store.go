package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"

	"github.com/AI2HU/gego/internal/jobs"
	"github.com/AI2HU/gego/internal/logger"
	"github.com/AI2HU/gego/internal/models"
)

const maxEtcdTxnOps = 124

type Store struct {
	client   *clientv3.Client
	keys     *Keys
	cfg      jobs.Config
	limiter  *RateLimiter
	isLeader bool
	leaderMu sync.RWMutex
	session  *concurrency.Session
	election *concurrency.Election
}

func NewStore(client *clientv3.Client, cfg jobs.Config) *Store {
	keys := NewKeys(cfg.Prefix)
	return &Store{
		client:  client,
		keys:    keys,
		cfg:     cfg,
		limiter: NewRateLimiter(client, keys, cfg.RequestsPerMinute),
	}
}

func (s *Store) RateLimiter() jobs.RateLimiter {
	return s.limiter
}

func (s *Store) Client() *clientv3.Client {
	return s.client
}

func (s *Store) Ping(ctx context.Context) error {
	_, err := s.client.Status(ctx, s.cfg.Endpoints[0])
	return err
}

func (s *Store) IsLeader() bool {
	s.leaderMu.RLock()
	defer s.leaderMu.RUnlock()
	return s.isLeader
}

func (s *Store) StartElection(ctx context.Context) error {
	session, err := concurrency.NewSession(s.client, concurrency.WithTTL(15))
	if err != nil {
		return fmt.Errorf("create election session: %w", err)
	}

	election := concurrency.NewElection(session, s.keys.LeaderLock())
	s.session = session
	s.election = election

	go func() {
		for {
			if err := election.Campaign(ctx, "scheduler"); err != nil {
				if ctx.Err() != nil {
					return
				}
				logger.Error("scheduler election campaign failed: %v", err)
				time.Sleep(time.Second)
				continue
			}

			s.leaderMu.Lock()
			s.isLeader = true
			s.leaderMu.Unlock()
			logger.Info("scheduler became leader")

			select {
			case <-session.Done():
				s.leaderMu.Lock()
				s.isLeader = false
				s.leaderMu.Unlock()
				logger.Info("scheduler lost leadership")
			case <-ctx.Done():
				s.leaderMu.Lock()
				s.isLeader = false
				s.leaderMu.Unlock()
				return
			}
		}
	}()

	return nil
}

func (s *Store) StopElection() {
	if s.election != nil {
		_ = s.election.Resign(context.Background())
	}
	if s.session != nil {
		_ = s.session.Close()
	}
	s.leaderMu.Lock()
	s.isLeader = false
	s.leaderMu.Unlock()
}

func (s *Store) CreateRun(ctx context.Context, run *models.ScheduleRun, jobList []*models.ScheduleJob, dedupKey string, dedupTTL time.Duration) error {
	if run == nil || len(jobList) == 0 {
		return fmt.Errorf("run and jobs are required")
	}

	now := time.Now().UTC()
	run.CreatedAt = now
	if run.Status == "" {
		run.Status = models.ScheduleRunStatusPending
	}

	runBytes, err := json.Marshal(run)
	if err != nil {
		return err
	}

	jobIDs := make([]string, len(jobList))
	ops := make([]clientv3.Op, 0, len(jobList)*2+4)
	ops = append(ops,
		clientv3.OpPut(s.keys.Run(run.ID), string(runBytes)),
		clientv3.OpPut(s.keys.RunIndex(now, run.ID), ""),
		clientv3.OpPut(s.keys.RunBySchedule(run.ScheduleID, now, run.ID), ""),
	)

	for i, job := range jobList {
		job.CreatedAt = now
		if job.Status == "" {
			job.Status = models.ScheduleJobStatusPending
		}
		if job.MaxAttempts == 0 {
			job.MaxAttempts = s.cfg.MaxRetries
		}
		pendingKey := s.keys.QueuePending(now.Add(time.Duration(i)*time.Nanosecond), job.ID)
		job.PendingKey = pendingKey
		jobIDs[i] = job.ID

		jobBytes, err := json.Marshal(job)
		if err != nil {
			return err
		}

		ops = append(ops,
			clientv3.OpPut(s.keys.JobData(job.ID), string(jobBytes)),
			clientv3.OpPut(pendingKey, job.ID),
		)
	}

	jobIDBytes, err := json.Marshal(jobIDs)
	if err != nil {
		return err
	}
	ops = append(ops, clientv3.OpPut(s.keys.RunJobIDs(run.ID), string(jobIDBytes)))

	if dedupKey != "" && dedupTTL > 0 {
		lease, err := s.client.Grant(ctx, int64(dedupTTL.Seconds()))
		if err != nil {
			return fmt.Errorf("grant dedup lease: %w", err)
		}
		ops = append(ops, clientv3.OpPut(dedupKey, run.ID, clientv3.WithLease(lease.ID)))
	}

	var committed bool
	defer func() {
		if committed {
			return
		}
		s.rollbackCreateRun(ctx, run, jobList, dedupKey)
	}()

	if err := s.commitBatchedOps(ctx, ops, dedupKey); err != nil {
		return err
	}
	committed = true
	return nil
}

func (s *Store) commitBatchedOps(ctx context.Context, ops []clientv3.Op, dedupKey string) error {
	for i := 0; i < len(ops); i += maxEtcdTxnOps {
		end := i + maxEtcdTxnOps
		if end > len(ops) {
			end = len(ops)
		}

		txn := s.client.Txn(ctx)
		if i == 0 && dedupKey != "" {
			txn = txn.If(clientv3.Compare(clientv3.CreateRevision(dedupKey), "=", 0))
		}

		resp, err := txn.Then(ops[i:end]...).Commit()
		if err != nil {
			return err
		}
		if i == 0 && dedupKey != "" && !resp.Succeeded {
			return fmt.Errorf("run already enqueued for this cron slot")
		}
	}
	return nil
}

func (s *Store) rollbackCreateRun(ctx context.Context, run *models.ScheduleRun, jobList []*models.ScheduleJob, dedupKey string) {
	ops := []clientv3.Op{
		clientv3.OpDelete(s.keys.Run(run.ID)),
		clientv3.OpDelete(s.keys.RunIndex(run.CreatedAt, run.ID)),
		clientv3.OpDelete(s.keys.RunJobIDs(run.ID)),
	}
	if run.ScheduleID != "" {
		ops = append(ops, clientv3.OpDelete(s.keys.RunBySchedule(run.ScheduleID, run.CreatedAt, run.ID)))
	}
	if dedupKey != "" {
		ops = append(ops, clientv3.OpDelete(dedupKey))
	}
	for _, job := range jobList {
		ops = append(ops, clientv3.OpDelete(s.keys.JobData(job.ID)))
		if job.PendingKey != "" {
			ops = append(ops, clientv3.OpDelete(job.PendingKey))
		}
	}
	_ = s.commitBatchedOps(ctx, ops, "")
}

func (s *Store) getJob(ctx context.Context, jobID string) (*models.ScheduleJob, error) {
	resp, err := s.client.Get(ctx, s.keys.JobData(jobID))
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}
	var job models.ScheduleJob
	if err := json.Unmarshal(resp.Kvs[0].Value, &job); err != nil {
		return nil, err
	}
	return &job, nil
}

func (s *Store) putJob(ctx context.Context, job *models.ScheduleJob) error {
	bytes, err := json.Marshal(job)
	if err != nil {
		return err
	}
	_, err = s.client.Put(ctx, s.keys.JobData(job.ID), string(bytes))
	return err
}

func (s *Store) getRun(ctx context.Context, runID string) (*models.ScheduleRun, error) {
	resp, err := s.client.Get(ctx, s.keys.Run(runID))
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, fmt.Errorf("run not found: %s", runID)
	}
	var run models.ScheduleRun
	if err := json.Unmarshal(resp.Kvs[0].Value, &run); err != nil {
		return nil, err
	}
	return &run, nil
}

func (s *Store) putRun(ctx context.Context, run *models.ScheduleRun) error {
	bytes, err := json.Marshal(run)
	if err != nil {
		return err
	}
	_, err = s.client.Put(ctx, s.keys.Run(run.ID), string(bytes))
	return err
}

func (s *Store) WatchPendingJobs(ctx context.Context) (<-chan *models.ScheduleJob, error) {
	ch := make(chan *models.ScheduleJob, 64)

	go func() {
		defer close(ch)
		prefix := s.keys.QueuePendingPrefix()
		var watchRev int64

		for {
			if ctx.Err() != nil {
				return
			}

			var watchChan clientv3.WatchChan
			if watchRev > 0 {
				watchChan = s.client.Watch(ctx, prefix, clientv3.WithPrefix(), clientv3.WithRev(watchRev))
			} else {
				if err := s.seedPending(ctx, ch); err != nil && ctx.Err() == nil {
					logger.Error("seed pending jobs: %v", err)
				}
				watchChan = s.client.Watch(ctx, prefix, clientv3.WithPrefix(), clientv3.WithRev(0))
			}

			for watchResp := range watchChan {
				if watchResp.Err() != nil {
					logger.Error("etcd watch error: %v", watchResp.Err())
					break
				}
				watchRev = watchResp.Header.Revision + 1
				for _, ev := range watchResp.Events {
					if ev.Type != clientv3.EventTypePut {
						continue
					}
					jobID := string(ev.Kv.Value)
					if jobID == "" {
						jobID = string(ev.Kv.Key[len(prefix):])
					}
					job, err := s.getJob(ctx, jobID)
					if err != nil {
						continue
					}
					job.PendingKey = string(ev.Kv.Key)
					select {
					case ch <- job:
					case <-ctx.Done():
						return
					}
				}
			}

			time.Sleep(time.Second)
		}
	}()

	return ch, nil
}

func (s *Store) seedPending(ctx context.Context, ch chan<- *models.ScheduleJob) error {
	resp, err := s.client.Get(ctx, s.keys.QueuePendingPrefix(), clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if err != nil {
		return err
	}
	for _, kv := range resp.Kvs {
		jobID := string(kv.Value)
		job, err := s.getJob(ctx, jobID)
		if err != nil {
			continue
		}
		job.PendingKey = string(kv.Key)
		select {
		case ch <- job:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func (s *Store) ClaimJob(ctx context.Context, job *models.ScheduleJob, workerID string, lease time.Duration) (*models.ScheduleJob, error) {
	if job.PendingKey == "" {
		return nil, fmt.Errorf("job missing pending key")
	}

	leaseResp, err := s.client.Grant(ctx, int64(lease.Seconds()))
	if err != nil {
		return nil, err
	}

	claimedKey := s.keys.QueueClaimed(job.ID)
	txn := s.client.Txn(ctx).
		If(clientv3.Compare(clientv3.ModRevision(job.PendingKey), ">", 0)).
		Then(
			clientv3.OpDelete(job.PendingKey),
			clientv3.OpPut(claimedKey, workerID, clientv3.WithLease(leaseResp.ID)),
		)

	txnResp, err := txn.Commit()
	if err != nil {
		return nil, err
	}
	if !txnResp.Succeeded {
		return nil, fmt.Errorf("job already claimed")
	}

	now := time.Now().UTC()
	job.WorkerID = workerID
	job.Status = models.ScheduleJobStatusClaimed
	job.Attempts++
	job.ClaimedAt = &now

	if err := s.putJob(ctx, job); err != nil {
		return nil, err
	}

	run, err := s.getRun(ctx, job.RunID)
	if err == nil && run.Status == models.ScheduleRunStatusPending {
		run.Status = models.ScheduleRunStatusRunning
		nowCopy := now
		run.StartedAt = &nowCopy
		_ = s.putRun(ctx, run)
	}

	return job, nil
}

func (s *Store) CompleteJob(ctx context.Context, jobID string, responseIDs []string) error {
	job, err := s.getJob(ctx, jobID)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	job.Status = models.ScheduleJobStatusCompleted
	job.ResponseIDs = append([]string(nil), responseIDs...)
	job.CompletedAt = &now
	job.Error = ""

	if err := s.putJob(ctx, job); err != nil {
		return err
	}

	_, _ = s.client.Delete(ctx, s.keys.QueueClaimed(jobID))

	run, err := s.UpdateRunProgress(ctx, job.RunID)
	if err != nil {
		return err
	}
	_ = run
	return nil
}

func (s *Store) FailJob(ctx context.Context, jobID string, jobErr error) error {
	job, err := s.getJob(ctx, jobID)
	if err != nil {
		return err
	}

	job.Error = jobErr.Error()
	_, _ = s.client.Delete(ctx, s.keys.QueueClaimed(jobID))

	if job.Attempts < job.MaxAttempts {
		job.Status = models.ScheduleJobStatusPending
		now := time.Now().UTC()
		pendingKey := s.keys.QueuePending(now, job.ID)
		job.PendingKey = pendingKey
		if err := s.putJob(ctx, job); err != nil {
			return err
		}
		_, err = s.client.Put(ctx, pendingKey, job.ID)
		return err
	}

	job.Status = models.ScheduleJobStatusDead
	if err := s.putJob(ctx, job); err != nil {
		return err
	}
	_, err = s.client.Put(ctx, s.keys.QueueDead(jobID), job.ID)
	if err != nil {
		return err
	}

	_, _ = s.UpdateRunProgress(ctx, job.RunID)
	return nil
}

func (s *Store) RetryJob(ctx context.Context, jobID string) error {
	job, err := s.getJob(ctx, jobID)
	if err != nil {
		return err
	}

	if job.Status != models.ScheduleJobStatusDead && job.Status != models.ScheduleJobStatusFailed {
		return fmt.Errorf("job is not retryable")
	}

	_, _ = s.client.Delete(ctx, s.keys.QueueDead(jobID))
	job.Status = models.ScheduleJobStatusPending
	job.Attempts = 0
	job.Error = ""
	job.WorkerID = ""
	job.ClaimedAt = nil
	job.CompletedAt = nil
	job.ResponseIDs = nil

	now := time.Now().UTC()
	pendingKey := s.keys.QueuePending(now, job.ID)
	job.PendingKey = pendingKey
	if err := s.putJob(ctx, job); err != nil {
		return err
	}
	_, err = s.client.Put(ctx, pendingKey, job.ID)
	if err != nil {
		return err
	}

	run, err := s.getRun(ctx, job.RunID)
	if err == nil && isTerminalRun(run.Status) {
		run.Status = models.ScheduleRunStatusRunning
		run.FinishedAt = nil
		_ = s.putRun(ctx, run)
	}
	return nil
}

func (s *Store) CancelRun(ctx context.Context, runID string) error {
	run, err := s.getRun(ctx, runID)
	if err != nil {
		return err
	}

	jobs, err := s.ListJobs(ctx, runID)
	if err != nil {
		return err
	}

	ops := []clientv3.Op{}
	for _, job := range jobs {
		if job.Status == models.ScheduleJobStatusPending && job.PendingKey != "" {
			ops = append(ops, clientv3.OpDelete(job.PendingKey))
			job.Status = models.ScheduleJobStatusFailed
			job.Error = "cancelled"
			bytes, _ := json.Marshal(job)
			ops = append(ops, clientv3.OpPut(s.keys.JobData(job.ID), string(bytes)))
		}
	}

	now := time.Now().UTC()
	run.Status = models.ScheduleRunStatusCancelled
	run.FinishedAt = &now
	runBytes, _ := json.Marshal(run)
	ops = append(ops, clientv3.OpPut(s.keys.Run(runID), string(runBytes)))

	if len(ops) > 0 {
		err = s.commitBatchedOps(ctx, ops, "")
	}
	return err
}

func (s *Store) ReclaimStaleJobs(ctx context.Context) (int, error) {
	resp, err := s.client.Get(ctx, s.keys.QueueClaimedPrefix(), clientv3.WithPrefix())
	if err != nil {
		return 0, err
	}

	reclaimed := 0
	for _, kv := range resp.Kvs {
		jobID := strings.TrimPrefix(string(kv.Key), s.keys.QueueClaimedPrefix())
		leaseID := clientv3.LeaseID(kv.Lease)

		if leaseID != 0 {
			leaseResp, err := s.client.TimeToLive(ctx, leaseID)
			if err == nil && leaseResp.TTL > 0 {
				continue
			}
		}

		job, err := s.getJob(ctx, jobID)
		if err != nil {
			_, _ = s.client.Delete(ctx, string(kv.Key))
			continue
		}

		_, _ = s.client.Delete(ctx, string(kv.Key))
		job.Status = models.ScheduleJobStatusPending
		job.WorkerID = ""
		now := time.Now().UTC()
		pendingKey := s.keys.QueuePending(now, job.ID)
		job.PendingKey = pendingKey
		_ = s.putJob(ctx, job)
		_, _ = s.client.Put(ctx, pendingKey, job.ID)
		reclaimed++
	}

	return reclaimed, nil
}

func (s *Store) UpdateRunProgress(ctx context.Context, runID string) (*models.ScheduleRun, error) {
	run, err := s.getRun(ctx, runID)
	if err != nil {
		return nil, err
	}

	jobs, err := s.ListJobs(ctx, runID)
	if err != nil {
		return nil, err
	}

	completed := 0
	failed := 0
	for _, job := range jobs {
		switch job.Status {
		case models.ScheduleJobStatusCompleted:
			completed++
		case models.ScheduleJobStatusDead, models.ScheduleJobStatusFailed:
			failed++
		}
	}

	run.CompletedJobs = completed
	run.FailedJobs = failed

	terminal := completed+failed >= run.TotalJobs && run.TotalJobs > 0
	if terminal {
		now := time.Now().UTC()
		run.FinishedAt = &now
		if failed > 0 && completed == 0 {
			run.Status = models.ScheduleRunStatusFailed
		} else if failed > 0 {
			run.Status = models.ScheduleRunStatusFailed
		} else {
			run.Status = models.ScheduleRunStatusCompleted
		}
	}

	if err := s.putRun(ctx, run); err != nil {
		return nil, err
	}
	return run, nil
}

func (s *Store) GetRun(ctx context.Context, runID string) (*models.ScheduleRun, error) {
	return s.getRun(ctx, runID)
}

func (s *Store) ListRuns(ctx context.Context, filter jobs.RunFilter) ([]*models.ScheduleRun, string, error) {
	prefix := s.keys.RunIndexPrefix()
	if filter.ScheduleID != "" {
		prefix = s.keys.RunBySchedulePrefix(filter.ScheduleID)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}

	opts := []clientv3.OpOption{
		clientv3.WithPrefix(),
		clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend),
		clientv3.WithLimit(int64(limit + 1)),
	}
	if filter.Cursor != "" {
		opts = append(opts, clientv3.WithRange(filter.Cursor))
	}

	resp, err := s.client.Get(ctx, prefix, opts...)
	if err != nil {
		return nil, "", err
	}

	runs := make([]*models.ScheduleRun, 0, len(resp.Kvs))
	var nextCursor string
	for i, kv := range resp.Kvs {
		if i >= limit {
			nextCursor = string(kv.Key)
			break
		}
		parts := strings.Split(string(kv.Key), "/")
		runID := parts[len(parts)-1]
		run, err := s.getRun(ctx, runID)
		if err != nil {
			continue
		}
		if filter.Status != "" && run.Status != filter.Status {
			continue
		}
		runs = append(runs, run)
	}

	return runs, nextCursor, nil
}

func (s *Store) ListJobs(ctx context.Context, runID string) ([]*models.ScheduleJob, error) {
	resp, err := s.client.Get(ctx, s.keys.RunJobIDs(runID))
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return []*models.ScheduleJob{}, nil
	}

	var jobIDs []string
	if err := json.Unmarshal(resp.Kvs[0].Value, &jobIDs); err != nil {
		return nil, err
	}

	jobs := make([]*models.ScheduleJob, 0, len(jobIDs))
	for _, jobID := range jobIDs {
		job, err := s.getJob(ctx, jobID)
		if err != nil {
			continue
		}
		jobs = append(jobs, job)
	}
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].CreatedAt.Before(jobs[j].CreatedAt)
	})
	return jobs, nil
}

func (s *Store) PendingJobCount(ctx context.Context) (int, error) {
	resp, err := s.client.Get(ctx, s.keys.QueuePendingPrefix(), clientv3.WithPrefix(), clientv3.WithCountOnly())
	if err != nil {
		return 0, err
	}
	return int(resp.Count), nil
}

func (s *Store) ActiveRunCount(ctx context.Context) (int, error) {
	resp, err := s.client.Get(ctx, s.keys.RunIndexPrefix(), clientv3.WithPrefix())
	if err != nil {
		return 0, err
	}

	count := 0
	for _, kv := range resp.Kvs {
		parts := strings.Split(string(kv.Key), "/")
		runID := parts[len(parts)-1]
		run, err := s.getRun(ctx, runID)
		if err != nil {
			continue
		}
		if run.Status == models.ScheduleRunStatusPending || run.Status == models.ScheduleRunStatusRunning {
			count++
		}
	}
	return count, nil
}

func (s *Store) RegisterWorker(ctx context.Context, workerID string, info *models.WorkerInfo, lease time.Duration) error {
	leaseResp, err := s.client.Grant(ctx, int64(lease.Seconds()))
	if err != nil {
		return err
	}
	bytes, err := json.Marshal(info)
	if err != nil {
		return err
	}
	_, err = s.client.Put(ctx, s.keys.Worker(workerID), string(bytes), clientv3.WithLease(leaseResp.ID))
	return err
}

func (s *Store) RefreshWorker(ctx context.Context, workerID string, info *models.WorkerInfo, lease time.Duration) error {
	return s.RegisterWorker(ctx, workerID, info, lease)
}

func (s *Store) UnregisterWorker(ctx context.Context, workerID string) error {
	_, err := s.client.Delete(ctx, s.keys.Worker(workerID))
	return err
}

func (s *Store) ListWorkers(ctx context.Context) ([]*models.WorkerInfo, error) {
	resp, err := s.client.Get(ctx, s.keys.WorkersPrefix(), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	workers := make([]*models.WorkerInfo, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var info models.WorkerInfo
		if err := json.Unmarshal(kv.Value, &info); err != nil {
			continue
		}
		workers = append(workers, &info)
	}
	return workers, nil
}

func (s *Store) CompactExpiredRuns(ctx context.Context, retention time.Duration) (int, error) {
	cutoff := time.Now().UTC().Add(-retention)
	resp, err := s.client.Get(ctx, s.keys.RunIndexPrefix(), clientv3.WithPrefix())
	if err != nil {
		return 0, err
	}

	removed := 0
	for _, kv := range resp.Kvs {
		parts := strings.Split(string(kv.Key), "/")
		if len(parts) < 2 {
			continue
		}
		tsStr := parts[len(parts)-2]
		tsNano, err := strconv.ParseInt(tsStr, 10, 64)
		if err != nil {
			continue
		}
		if time.Unix(0, tsNano).After(cutoff) {
			continue
		}

		runID := parts[len(parts)-1]
		run, err := s.getRun(ctx, runID)
		if err != nil || !isTerminalRun(run.Status) {
			continue
		}

		jobs, _ := s.ListJobs(ctx, runID)
		ops := []clientv3.Op{
			clientv3.OpDelete(s.keys.Run(runID)),
			clientv3.OpDelete(string(kv.Key)),
			clientv3.OpDelete(s.keys.RunJobIDs(runID)),
		}
		if run.ScheduleID != "" {
			ops = append(ops, clientv3.OpDelete(s.keys.RunBySchedule(run.ScheduleID, run.CreatedAt, runID)))
		}
		for _, job := range jobs {
			ops = append(ops,
				clientv3.OpDelete(s.keys.JobData(job.ID)),
				clientv3.OpDelete(s.keys.QueueDead(job.ID)),
			)
		}
		_ = s.commitBatchedOps(ctx, ops, "")
		removed++
	}
	return removed, nil
}

func isTerminalRun(status models.ScheduleRunStatus) bool {
	switch status {
	case models.ScheduleRunStatusCompleted, models.ScheduleRunStatusFailed, models.ScheduleRunStatusCancelled:
		return true
	default:
		return false
	}
}
