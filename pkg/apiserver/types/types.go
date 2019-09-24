package types

import (
	"github.com/juju/errors"
)

// Job represents a chaos job
type Job struct {
	ID        int64       `json:"id" db:"id"`
	Pods      []string    `json:"pods"`
	JobType   string      `json:"jobType" db:"job_type"`
	EventType string      `json:"eventType" db:"event_type"`
	Resource  interface{} `json:"resource" db:"resource"`
	Ctime     string      `json:"create_time" db:"create_time"`
}

// JobPodRelation represents a relation between job and pod in database
type JobPodRelation struct {
	JobID int64  `db:"job_id"`
	Pod   string `db:"pod"`
}

// JobPodJoinSelect represents a select result of job join job_pod
type JobPodJoinSelect struct {
	Job
	PodsStr string `db:"pods_str"`
}

// Verify will verify whether a job is valid
func (job *Job) Verify() error {
	if job.EventType != "start" &&
		job.EventType != "oneshot" &&
		job.EventType != "end" {
		return errors.New("unknown event type")
	}

	return nil
}
