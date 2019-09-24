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

// MysqlClient is a client for querying mysql
type MysqlClient struct {
	db *sqlx.DB
}

// NewMysqlClient will create a mysql client
func NewMysqlClient(dataSource string) (*MysqlClient, error) {
	log.Infof("connecting to %s", dataSource)
	db, err := sqlx.Open("mysql", dataSource)
	if err != nil {
		return nil, errors.Trace(err)
	}
	log.Info("database connected")

	return &MysqlClient{
		db,
	}, nil
}

// CreateJob will insert a job into database
func (m *MysqlClient) CreateJob(job *types.Job) error {
	t := time.Now().Format(util.TimeFormat)
	job.Ctime = t

	tx, err := m.db.Beginx()
	if err != nil {
		log.Error(err)
		return errors.Trace(err)
	}

	resource, err := json.Marshal(job.Resource)
	if err != nil {
		log.Error(err)
		return errors.Trace(err)
	}
	result, err := tx.NamedExec(jobInsert, map[string]interface{}{
		"event_type":  job.EventType,
		"resource":    string(resource),
		"create_time": job.Ctime,
		"job_type":    job.JobType,
	})
	if err != nil {
		tx.Rollback()
		log.Error(err)
		return errors.Trace(err)
	}

	job.ID, err = result.LastInsertId()
	if err != nil {
		tx.Rollback()
		log.Error(err)
		return errors.Trace(err)
	}

	_, err = tx.NamedExec(jobPodInsert, generatePodRelation(job))
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

// GetJobs will select jobs from database
func (m *MysqlClient) GetJobs(fs *filter.Filters) ([]*types.Job, error) {
	filtersSQL, err := filter.GenSQL(fs)
	if err != nil {
		return nil, errors.Trace(err)
	}

	rows, err := m.db.Queryx(fmt.Sprintf(jobSelect, " WHERE "+filtersSQL))
	if err != nil {
		log.Error(err)
		return nil, errors.Trace(err)
	}

	var jobs []*types.Job
	for rows.Next() {
		job := new(types.JobPodJoinSelect)
		rows.StructScan(&job)

		resource, ok := job.Resource.([]byte)
		if !ok {
			return nil, errors.New("resource is not []byte")
		}
		json.Unmarshal(resource, &job.Resource)

		job.Pods = strings.Split(job.PodsStr, ",")
		jobs = append(jobs, &job.Job)
	}

	return jobs, nil
}

func generatePodRelation(job *types.Job) []types.JobPodRelation {
	var list []types.JobPodRelation

	for _, pod := range job.Pods {
		list = append(list, types.JobPodRelation{
			JobID: job.ID,
			Pod:   pod,
		})
	}

	return list
}

const jobInsert = `
	INSERT INTO job (
		event_type,
		resource,
		job_type,
		create_time
	) VALUES (
		:event_type,
		:resource,
		:job_type,
		:create_time
	)
`

const jobPodInsert = `
	INSERT INTO job_pod (
		job_id,
		pod
	) VALUES (
		:job_id,
		:pod
	)
`

const jobSelect = `
  SELECT id,event_type,resource,job_type,create_time,GROUP_CONCAT(pod separator ',') AS pods_str FROM job JOIN job_pod ON id=job_id %s GROUP BY id
`
