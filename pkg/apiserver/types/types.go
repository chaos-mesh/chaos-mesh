package types

import (
	"github.com/juju/errors"
)

type Pod struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// Task represents a chaos task
type Task struct {
	ID        int64       `json:"id" db:"id"`
	Pods      []Pod       `json:"pods"`
	TaskType  string      `json:"taskType" db:"task_type"`
	EventType string      `json:"eventType" db:"event_type"`
	Resource  interface{} `json:"resource" db:"resource"`
	Ctime     string      `json:"create_time" db:"create_time"`
}

// TaskPodRelation represents a relation between task and pod in database
type TaskPodRelation struct {
	TaskID       int64  `db:"task_id"`
	PodName      string `db:"pod_name"`
	PodNamespace string `db:"pod_namespace"`
}

// TaskPodJoinSelect represents a select result of task join task_pod
type TaskPodJoinSelect struct {
	Task
	PodsNameStr      string `db:"pods_name_str"`
	PodsNamespaceStr string `db:"pods_namespace_str"`
}

// Verify will verify whether a task is valid
func (task *Task) Verify() error {
	if task.EventType != "start" &&
		task.EventType != "oneshot" &&
		task.EventType != "end" {
		return errors.New("unknown event type")
	}

	return nil
}
