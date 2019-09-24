package filter

import (
	"fmt"
	"strings"

	"github.com/juju/errors"
)

// Filters is a slice of Filter
type Filters = []Filter

// Filter represent a filter
type Filter struct {
	FilterType string      `json:"filterType"`
	Content    interface{} `json:"content"`
}

// GenSQL will generate sql for Filters (concat with AND)
func GenSQL(fs *Filters) (string, error) {
	if fs == nil {
		return "TRUE", nil
	}

	filterList := []Filter(*fs)

	filterSQLs := []string{"TRUE"}

	for _, filter := range filterList {
		sql, err := filter.GenSQL()
		if err != nil {
			return "", errors.Trace(err)
		}
		filterSQLs = append(filterSQLs, sql)
	}

	return strings.Join(filterSQLs, " AND "), nil
}

// GenSQL will generate sql for filter
func (filter *Filter) GenSQL() (string, error) {
	// FIXME: There are SQL Injection bugs here.
	switch filter.FilterType {
	case "pods":
		pods, ok := filter.Content.([]interface{})
		if !ok {
			return "", errors.New("content of pods filter is not []interface{}")
		}

		existSQL := []string{"TRUE"}
		for _, pod := range pods {
			existSQL = append(existSQL, fmt.Sprintf(existPodSQL, pod))
		}

		return strings.Join(existSQL, " AND "), nil

	case "type":
		jobType, ok := filter.Content.(string)
		if !ok {
			return "", errors.New("content of type filter is not string")
		}

		return fmt.Sprintf(`job_type="%s"`, jobType), nil

	default:
		return "", errors.New(fmt.Sprintf("unsupported filter type: %s", filter.FilterType))
	}
}

const existPodSQL = `EXISTS (SELECT 1 FROM job_pod WHERE job_id=id AND pod="%s")`
