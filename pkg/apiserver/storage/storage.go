package storage

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/juju/errors"
	"github.com/ngaut/log"
	"github.com/pingcap/chaos-operator/pkg/apiserver/filter"
	"github.com/pingcap/chaos-operator/pkg/apiserver/types"
	"github.com/pingcap/chaos-operator/util"
)

// SqlClient is a client for querying mysql
type SqlClient struct {
	db *sqlx.DB
}

const databaseSchema = `
CREATE TABLE IF NOT EXISTS task (
  id INTEGER PRIMARY KEY ASC, 
  event_type VARCHAR(20) NOT NULL,
  task_type VARCHAR(255) NOT NULL,
  resource TEXT,
  create_time DATETIME
);

CREATE TABLE IF NOT EXISTS task_pod (
  task_id INTEGER NOT NULL,
  pod_namespace VARCHAR(255) NOT NULL,
  pod_name VARCHAR(255) NOT NULL,
  FOREIGN KEY (task_id) REFERENCES task(id)
);
`

// NewMysqlClient will create a mysql client
func NewSqlClient() (*SqlClient, error) {
	log.Infof("connecting to sqlite")
	db, err := sqlx.Open("sqlite3", "/data/api-server.db")
	db.MustExec(databaseSchema)

	if err != nil {
		return nil, errors.Trace(err)
	}
	log.Info("database connected")

	return &SqlClient{
		db,
	}, nil
}

// CreateTask will insert a task into database
func (m *SqlClient) CreateTask(task *types.Task) error {
	t := time.Now().Format(util.TimeFormat)
	task.Ctime = t

	tx, err := m.db.Beginx()
	if err != nil {
		log.Error(err)
		return errors.Trace(err)
	}

	resource, err := json.Marshal(task.Resource)
	if err != nil {
		log.Error(err)
		return errors.Trace(err)
	}
	result, err := tx.NamedExec(taskInsert, map[string]interface{}{
		"event_type":  task.EventType,
		"resource":    string(resource),
		"create_time": task.Ctime,
		"task_type":   task.TaskType,
	})
	if err != nil {
		tx.Rollback()
		log.Error(err)
		return errors.Trace(err)
	}

	task.ID, err = result.LastInsertId()
	if err != nil {
		tx.Rollback()
		log.Error(err)
		return errors.Trace(err)
	}

	relation := generatePodRelation(task)
	_, err = tx.NamedExec(taskPodInsert, relation)
	if err != nil {
		tx.Rollback()
		log.Error(err)
		return errors.Trace(err)
	}

	if err = tx.Commit(); err != nil {
		log.Error(err)
		return errors.Trace(err)
	}

	return nil
}

// GetTasks will select tasks from database
func (m *SqlClient) GetTasks(filter *filter.Filter) ([]*types.Task, error) {
	filtersSQL, err := filter.GenSQL()
	if err != nil {
		return nil, errors.Trace(err)
	}

	rows, err := m.db.Queryx(fmt.Sprintf(taskSelect, " WHERE "+filtersSQL))
	if err != nil {
		log.Error(err)
		return nil, errors.Trace(err)
	}

	var tasks []*types.Task
	for rows.Next() {
		task := new(types.TaskPodJoinSelect)
		rows.StructScan(&task)

		resource, ok := task.Resource.(string)
		if !ok {
			return nil, errors.Errorf("resource is not string but %T", task.Resource)
		}
		json.Unmarshal([]byte(resource), &task.Resource)

		namespaces := strings.Split(task.PodsNamespaceStr, ",")
		names := strings.Split(task.PodsNameStr, ",")
		var pods []types.Pod
		for index, namespace := range namespaces {
			pods = append(pods, types.Pod{
				Name:      names[index],
				Namespace: namespace,
			})
		}

		task.Pods = pods
		tasks = append(tasks, &task.Task)
	}

	return tasks, nil
}

func generatePodRelation(task *types.Task) []types.TaskPodRelation {
	var list []types.TaskPodRelation

	for _, pod := range task.Pods {
		list = append(list, types.TaskPodRelation{
			TaskID:       task.ID,
			PodName:      pod.Name,
			PodNamespace: pod.Namespace,
		})
	}

	return list
}

const taskInsert = `
	INSERT INTO task (
		event_type,
		resource,
		task_type,
		create_time
	) VALUES (
		:event_type,
		:resource,
		:task_type,
		:create_time
	)
`

const taskPodInsert = `
	INSERT INTO task_pod (
		task_id,
		pod_name,
		pod_namespace
	) VALUES (
		:task_id,
		:pod_name,
		:pod_namespace
	)
`

const taskSelect = `
  SELECT id,event_type,resource,task_type,create_time,GROUP_CONCAT(pod_name) AS pods_name_str,GROUP_CONCAT(pod_namespace) AS pods_namespace_str FROM task JOIN task_pod ON id=task_id %s GROUP BY id
`
