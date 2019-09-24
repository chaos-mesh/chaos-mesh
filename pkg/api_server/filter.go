package api_server

import (
	"strings"

	"github.com/juju/errors"
)

type Filters = []Filter

type Filter struct {
	FilterType string      `json:"filterType"`
	Content    interface{} `json:"content"`
}

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

func (filter *Filter) GenSQL() (string, error) {
	return "TRUE", nil
}
