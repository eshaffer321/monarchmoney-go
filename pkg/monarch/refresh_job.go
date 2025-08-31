package monarch

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
)

// RefreshStatus represents the status of a refresh job
type RefreshStatus string

const (
	RefreshStatusPending    RefreshStatus = "pending"
	RefreshStatusInProgress RefreshStatus = "in_progress"
	RefreshStatusCompleted  RefreshStatus = "completed"
	RefreshStatusFailed     RefreshStatus = "failed"
	RefreshStatusCancelled  RefreshStatus = "cancelled"
	RefreshStatusTimeout    RefreshStatus = "timeout"
)

// refreshJob implements the RefreshJob interface with proper status tracking
type refreshJob struct {
	client     *Client
	accountIDs []string
	id         string
	startTime  time.Time
	endTime    *time.Time

	// Status tracking
	status     atomic.Value // RefreshStatus
	lastCheck  time.Time
	checkCount int32

	// Cancellation
	cancelFunc context.CancelFunc
	cancelled  atomic.Bool

	// Error tracking
	lastError error
	errorLock sync.RWMutex

	// Progress tracking
	progress    map[string]bool // accountID -> completed
	progressMux sync.RWMutex
}

// newRefreshJob creates a new refresh job with proper initialization
func newRefreshJob(client *Client, accountIDs []string) *refreshJob {
	job := &refreshJob{
		client:     client,
		accountIDs: accountIDs,
		id:         fmt.Sprintf("refresh-%d", time.Now().Unix()),
		startTime:  time.Now(),
		progress:   make(map[string]bool),
	}

	// Initialize progress map
	for _, id := range accountIDs {
		job.progress[id] = false
	}

	// Set initial status
	job.status.Store(RefreshStatusPending)

	return job
}

// ID returns the job ID
func (j *refreshJob) ID() string {
	return j.id
}

// Status returns the current status
func (j *refreshJob) Status() RefreshStatus {
	return j.status.Load().(RefreshStatus)
}

// Wait waits for the job to complete with timeout and proper cancellation
func (j *refreshJob) Wait(ctx context.Context, timeout time.Duration) error {
	// Create cancellable context
	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Store cancel function for potential cancellation
	j.cancelFunc = cancel

	// Update status
	j.status.Store(RefreshStatusInProgress)

	// Polling configuration
	const (
		initialInterval = 1 * time.Second
		maxInterval     = 5 * time.Second
		backoffFactor   = 1.5
	)

	currentInterval := initialInterval
	ticker := time.NewTicker(currentInterval)
	defer ticker.Stop()

	for {
		select {
		case <-waitCtx.Done():
			// Check if it was cancelled vs timeout
			if j.cancelled.Load() {
				j.status.Store(RefreshStatusCancelled)
				return errors.New("refresh job was cancelled")
			}

			j.status.Store(RefreshStatusTimeout)
			now := time.Now()
			j.endTime = &now
			return ErrRefreshTimeout

		case <-ticker.C:
			// Check completion status
			complete, err := j.checkStatus(waitCtx)

			if err != nil {
				j.setError(err)

				// Check if error is retryable
				if !IsRetryable(err) {
					j.status.Store(RefreshStatusFailed)
					now := time.Now()
					j.endTime = &now
					return err
				}

				// Continue polling on retryable errors
				continue
			}

			if complete {
				j.status.Store(RefreshStatusCompleted)
				now := time.Now()
				j.endTime = &now
				return nil
			}

			// Implement exponential backoff
			atomic.AddInt32(&j.checkCount, 1)
			if j.checkCount%3 == 0 && currentInterval < maxInterval {
				currentInterval = time.Duration(float64(currentInterval) * backoffFactor)
				if currentInterval > maxInterval {
					currentInterval = maxInterval
				}
				ticker.Reset(currentInterval)
			}
		}
	}
}

// IsComplete checks if the job is complete
func (j *refreshJob) IsComplete(ctx context.Context) (bool, error) {
	status := j.Status()

	switch status {
	case RefreshStatusCompleted:
		return true, nil
	case RefreshStatusFailed, RefreshStatusCancelled, RefreshStatusTimeout:
		return false, j.getError()
	case RefreshStatusPending:
		// Start the job if not started
		return false, nil
	case RefreshStatusInProgress:
		// Check actual status
		return j.checkStatus(ctx)
	default:
		return false, fmt.Errorf("unknown status: %s", status)
	}
}

// Cancel cancels the job
func (j *refreshJob) Cancel(ctx context.Context) error {
	// Mark as cancelled
	j.cancelled.Store(true)

	// Call cancel function if available
	if j.cancelFunc != nil {
		j.cancelFunc()
	}

	// Update status
	j.status.Store(RefreshStatusCancelled)
	now := time.Now()
	j.endTime = &now

	return nil
}

// GetProgress returns the progress of individual accounts
func (j *refreshJob) GetProgress() map[string]bool {
	j.progressMux.RLock()
	defer j.progressMux.RUnlock()

	// Return a copy to avoid race conditions
	progress := make(map[string]bool)
	for k, v := range j.progress {
		progress[k] = v
	}
	return progress
}

// GetMetrics returns job metrics
func (j *refreshJob) GetMetrics() RefreshJobMetrics {
	status := j.Status()
	duration := time.Since(j.startTime)
	if j.endTime != nil {
		duration = j.endTime.Sub(j.startTime)
	}

	progress := j.GetProgress()
	completedCount := 0
	for _, completed := range progress {
		if completed {
			completedCount++
		}
	}

	return RefreshJobMetrics{
		ID:             j.id,
		Status:         string(status),
		StartTime:      j.startTime,
		EndTime:        j.endTime,
		Duration:       duration,
		AccountCount:   len(j.accountIDs),
		CompletedCount: completedCount,
		CheckCount:     int(atomic.LoadInt32(&j.checkCount)),
		LastCheck:      j.lastCheck,
		LastError:      j.getError(),
	}
}

// checkStatus checks the actual refresh status from the API
func (j *refreshJob) checkStatus(ctx context.Context) (bool, error) {
	j.lastCheck = time.Now()

	// Query account status
	query := j.client.loadQuery("accounts/refresh_status.graphql")

	variables := map[string]interface{}{
		"accountIds": j.accountIDs,
	}

	var result struct {
		Accounts []struct {
			ID         string     `json:"id"`
			Syncing    bool       `json:"syncing"`
			LastSynced *time.Time `json:"lastSyncedAt"`
			Credential *struct {
				UpdateRequired                 bool       `json:"updateRequired"`
				DisconnectedFromDataProviderAt *time.Time `json:"disconnectedFromDataProviderAt"`
			} `json:"credential"`
		} `json:"accounts"`
	}

	if err := j.client.executeGraphQL(ctx, query, variables, &result); err != nil {
		return false, errors.Wrap(err, "failed to check refresh status")
	}

	// Update progress for each account
	j.progressMux.Lock()
	allComplete := true

	for _, account := range result.Accounts {
		// Check if account is still syncing
		if account.Syncing {
			j.progress[account.ID] = false
			allComplete = false
			continue
		}

		// Check if credential needs update
		if account.Credential != nil && account.Credential.UpdateRequired {
			j.progress[account.ID] = false
			allComplete = false
			continue
		}

		// Check if sync happened after job start
		if account.LastSynced != nil && account.LastSynced.After(j.startTime) {
			j.progress[account.ID] = true
		} else {
			// Account not syncing but hasn't been updated
			j.progress[account.ID] = false
			allComplete = false
		}
	}

	j.progressMux.Unlock()

	return allComplete, nil
}

// setError sets the last error
func (j *refreshJob) setError(err error) {
	j.errorLock.Lock()
	defer j.errorLock.Unlock()
	j.lastError = err
}

// getError gets the last error
func (j *refreshJob) getError() error {
	j.errorLock.RLock()
	defer j.errorLock.RUnlock()
	return j.lastError
}

// RefreshJobMetrics contains metrics about a refresh job
type RefreshJobMetrics struct {
	ID             string        `json:"id"`
	Status         string        `json:"status"`
	StartTime      time.Time     `json:"startTime"`
	EndTime        *time.Time    `json:"endTime,omitempty"`
	Duration       time.Duration `json:"duration"`
	AccountCount   int           `json:"accountCount"`
	CompletedCount int           `json:"completedCount"`
	CheckCount     int           `json:"checkCount"`
	LastCheck      time.Time     `json:"lastCheck"`
	LastError      error         `json:"lastError,omitempty"`
}

// RefreshJobManager manages multiple refresh jobs
type RefreshJobManager struct {
	jobs map[string]*refreshJob
	mu   sync.RWMutex
}

// NewRefreshJobManager creates a new refresh job manager
func NewRefreshJobManager() *RefreshJobManager {
	return &RefreshJobManager{
		jobs: make(map[string]*refreshJob),
	}
}

// AddJob adds a job to the manager
func (m *RefreshJobManager) AddJob(job *refreshJob) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.jobs[job.ID()] = job
}

// GetJob retrieves a job by ID
func (m *RefreshJobManager) GetJob(id string) (*refreshJob, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	job, exists := m.jobs[id]
	return job, exists
}

// ListJobs lists all jobs
func (m *RefreshJobManager) ListJobs() []*refreshJob {
	m.mu.RLock()
	defer m.mu.RUnlock()

	jobs := make([]*refreshJob, 0, len(m.jobs))
	for _, job := range m.jobs {
		jobs = append(jobs, job)
	}
	return jobs
}

// CleanupCompleted removes completed jobs older than the specified duration
func (m *RefreshJobManager) CleanupCompleted(olderThan time.Duration) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	removed := 0

	for id, job := range m.jobs {
		status := job.Status()
		if (status == RefreshStatusCompleted || status == RefreshStatusFailed ||
			status == RefreshStatusCancelled || status == RefreshStatusTimeout) &&
			job.endTime != nil && now.Sub(*job.endTime) > olderThan {
			delete(m.jobs, id)
			removed++
		}
	}

	return removed
}
