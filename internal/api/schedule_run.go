package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/AI2HU/gego/internal/jobs"
	"github.com/AI2HU/gego/internal/models"
)

func (s *Server) listScheduleRuns(c *gin.Context) {
	ctx, cancel := etcdRequestContext(c.Request.Context())
	defer cancel()

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	filter := jobs.RunFilter{
		ScheduleID: c.Query("schedule_id"),
		Cursor:     c.Query("cursor"),
		Limit:      limit,
	}
	if status := c.Query("status"); status != "" {
		filter.Status = models.ScheduleRunStatus(status)
	}

	runs, nextCursor, err := s.jobStore.ListRuns(ctx, filter)
	if err != nil {
		s.errorResponse(c, http.StatusServiceUnavailable, "etcd unavailable: "+err.Error())
		return
	}

	responses := make([]models.ScheduleRunResponse, len(runs))
	for i, run := range runs {
		responses[i] = toScheduleRunResponse(run)
	}

	s.successResponse(c, models.ScheduleRunListResponse{
		Data:       responses,
		NextCursor: nextCursor,
	})
}

func (s *Server) getScheduleRun(c *gin.Context) {
	ctx, cancel := etcdRequestContext(c.Request.Context())
	defer cancel()

	run, err := s.jobStore.GetRun(ctx, c.Param("id"))
	if err != nil {
		s.errorResponse(c, http.StatusNotFound, "Schedule run not found: "+err.Error())
		return
	}

	jobList, err := s.jobStore.ListJobs(ctx, run.ID)
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to list jobs: "+err.Error())
		return
	}

	resp := toScheduleRunResponse(run)
	jobsResp := make([]models.ScheduleJobResponse, len(jobList))
	for i, job := range jobList {
		jobsResp[i] = toScheduleJobResponse(job)
	}

	s.successResponse(c, models.ScheduleRunDetailResponse{
		Run:  resp,
		Jobs: jobsResp,
	})
}

func (s *Server) listScheduleRunJobs(c *gin.Context) {
	ctx, cancel := etcdRequestContext(c.Request.Context())
	defer cancel()

	jobList, err := s.jobStore.ListJobs(ctx, c.Param("id"))
	if err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to list jobs: "+err.Error())
		return
	}

	responses := make([]models.ScheduleJobResponse, len(jobList))
	for i, job := range jobList {
		responses[i] = toScheduleJobResponse(job)
	}
	s.successResponse(c, responses)
}

func (s *Server) cancelScheduleRun(c *gin.Context) {
	ctx, cancel := etcdRequestContext(c.Request.Context())
	defer cancel()

	if err := s.jobStore.CancelRun(ctx, c.Param("id")); err != nil {
		s.errorResponse(c, http.StatusInternalServerError, "Failed to cancel run: "+err.Error())
		return
	}
	s.successResponse(c, gin.H{"message": "run cancelled"})
}

func (s *Server) retryScheduleJob(c *gin.Context) {
	ctx, cancel := etcdRequestContext(c.Request.Context())
	defer cancel()

	if err := s.jobStore.RetryJob(ctx, c.Param("job_id")); err != nil {
		s.errorResponse(c, http.StatusBadRequest, "Failed to retry job: "+err.Error())
		return
	}
	s.successResponse(c, gin.H{"message": "job requeued"})
}

func toScheduleRunResponse(run *models.ScheduleRun) models.ScheduleRunResponse {
	return models.ScheduleRunResponse{
		ID:            run.ID,
		ScheduleID:    run.ScheduleID,
		Trigger:       string(run.Trigger),
		Status:        string(run.Status),
		TotalJobs:     run.TotalJobs,
		CompletedJobs: run.CompletedJobs,
		FailedJobs:    run.FailedJobs,
		CreatedAt:     run.CreatedAt,
		StartedAt:     run.StartedAt,
		FinishedAt:    run.FinishedAt,
	}
}

func toScheduleJobResponse(job *models.ScheduleJob) models.ScheduleJobResponse {
	return models.ScheduleJobResponse{
		ID:          job.ID,
		RunID:       job.RunID,
		ScheduleID:  job.ScheduleID,
		PromptID:    job.PromptID,
		LLMID:       job.LLMID,
		Provider:    job.Provider,
		Temperature: job.Temperature,
		Status:      string(job.Status),
		Attempts:    job.Attempts,
		MaxAttempts: job.MaxAttempts,
		WorkerID:    job.WorkerID,
		ResponseID:  job.ResponseID,
		Error:       job.Error,
		CreatedAt:   job.CreatedAt,
		ClaimedAt:   job.ClaimedAt,
		CompletedAt: job.CompletedAt,
	}
}
