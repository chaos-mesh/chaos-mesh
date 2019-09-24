package api_server

import (
	"github.com/juju/errors"
)

type Job struct {
	ID        int64       `json:"id" db:"id"`
	Pods      []string    `json:"pods"`
	JobType   string      `json:"jobType" db:"job_type"`
	EventType string      `json:"eventType" db:"event_type"`
	Resource  interface{} `json:"resource" db:"resource"`
	Ctime     string      `json:"create_time" db:"create_time"`
}

type JobPodRelation struct {
	JobID int64  `db:"job_id"`
	Pod   string `db:"pod"`
}

type JobPodJoinSelect struct {
	Job
	PodsStr string `db:"pods_str"`
}

func (job *Job) Verify() error {
	if job.EventType != "start" &&
		job.EventType != "oneshot" &&
		job.EventType != "end" {
		return errors.New("unknown event type")
	}

	return nil
}
