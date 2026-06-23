package etcd

import (
	"fmt"
	"path"
	"strings"
	"time"
)

type Keys struct {
	prefix string
}

func NewKeys(prefix string) *Keys {
	prefix = strings.TrimSuffix(prefix, "/")
	if prefix == "" {
		prefix = "/gego"
	}
	return &Keys{prefix: prefix}
}

func (k *Keys) Run(runID string) string {
	return path.Join(k.prefix, "runs", runID)
}

func (k *Keys) RunIndex(createdAt time.Time, runID string) string {
	return path.Join(k.prefix, "runs", "index", fmt.Sprintf("%020d", createdAt.UnixNano()), runID)
}

func (k *Keys) RunBySchedule(scheduleID string, createdAt time.Time, runID string) string {
	return path.Join(k.prefix, "runs", "by-schedule", scheduleID, fmt.Sprintf("%020d", createdAt.UnixNano()), runID)
}

func (k *Keys) RunJobIDs(runID string) string {
	return path.Join(k.prefix, "runs", runID, "job-ids")
}

func (k *Keys) JobData(jobID string) string {
	return path.Join(k.prefix, "jobs", "data", jobID)
}

func (k *Keys) QueuePending(createdAt time.Time, jobID string) string {
	return path.Join(k.prefix, "queue", "pending", fmt.Sprintf("%020d", createdAt.UnixNano()), jobID)
}

func (k *Keys) QueuePendingPrefix() string {
	return path.Join(k.prefix, "queue", "pending") + "/"
}

func (k *Keys) QueueClaimed(jobID string) string {
	return path.Join(k.prefix, "queue", "claimed", jobID)
}

func (k *Keys) QueueClaimedPrefix() string {
	return path.Join(k.prefix, "queue", "claimed") + "/"
}

func (k *Keys) QueueDead(jobID string) string {
	return path.Join(k.prefix, "queue", "dead", jobID)
}

func (k *Keys) Dedup(scheduleID, cronSlot string) string {
	return path.Join(k.prefix, "dedup", scheduleID, cronSlot)
}

func (k *Keys) RateLimit(provider string) string {
	return path.Join(k.prefix, "ratelimit", provider, "tokens")
}

func (k *Keys) Worker(workerID string) string {
	return path.Join(k.prefix, "workers", workerID)
}

func (k *Keys) WorkersPrefix() string {
	return path.Join(k.prefix, "workers") + "/"
}

func (k *Keys) LeaderLock() string {
	return path.Join(k.prefix, "locks", "scheduler-leader")
}

func (k *Keys) RunIndexPrefix() string {
	return path.Join(k.prefix, "runs", "index") + "/"
}

func (k *Keys) RunBySchedulePrefix(scheduleID string) string {
	return path.Join(k.prefix, "runs", "by-schedule", scheduleID) + "/"
}

func (k *Keys) Prefix() string {
	return k.prefix
}
