package filter

import (
	"fmt"
	"strings"

	"github.com/juju/errors"
)

// Filters is a slice of Filter
// Filter represent a filter
type Filter struct {
	FilterType string      `json:"filterType"`
	Content    interface{} `json:"content"`
}

// GenSQL will generate sql for filter
func (filter *Filter) GenSQL() (string, error) {
	// FIXME: There are SQL Injection bugs here.
	switch filter.FilterType {
	case "pod_name":
		pods, ok := filter.Content.([]interface{})
		if !ok {
			return "", errors.New("content of pods filter is not []interface{}")
		}

		existSQL := []string{"TRUE"}
		for _, pod := range pods {
			existSQL = append(existSQL, fmt.Sprintf(existPodSQL, pod))
		}

		return strings.Join(existSQL, " AND "), nil

	case "namespace":
		namespaces, ok := filter.Content.([]interface{})
		if !ok {
			return "", errors.New("content of pod namespace filter is not []interface{}")
		}

		existSQL := []string{"TRUE"}
		for _, namespace := range namespaces {
			existSQL = append(existSQL, fmt.Sprintf(existNamespaceSQL, namespace))
		}

		return strings.Join(existSQL, " AND "), nil

	case "and":
		filters, ok := filter.Content.([]interface{})
		if !ok {
			return "", errors.New("content of and filter is not []interface{}")
		}

		sqls := []string{"TRUE"}
		fs, err := checkAndConvertListOfFilter(filters)
		if err != nil {
			return "", errors.New("parse content of \"and\" filter failed")
		}
		for _, f := range *fs {
			sql, err := f.GenSQL()

			if err != nil {
				return "", err
			}
			sqls = append(sqls, fmt.Sprintf("(%s)", sql))
		}

		return strings.Join(sqls, " AND "), nil

	case "or":
		filters, ok := filter.Content.([]interface{})
		if !ok {
			return "", errors.New("content of and filter is not []interface{}")
		}

		sqls := []string{"TRUE"}
		fs, err := checkAndConvertListOfFilter(filters)
		if err != nil {
			return "", errors.New("parse content of \"and\" filter failed")
		}
		for _, f := range *fs {
			sql, err := f.GenSQL()

			if err != nil {
				return "", err
			}
			sqls = append(sqls, fmt.Sprintf("(%s)", sql))
		}

		return strings.Join(sqls, " OR "), nil

	case "type":
		taskType, ok := filter.Content.(string)
		if !ok {
			return "", errors.New("content of type filter is not string")
		}

		return fmt.Sprintf(`task_type="%s"`, taskType), nil

	default:
		return "", errors.New(fmt.Sprintf("unsupported filter type: %s", filter.FilterType))
	}
}

func checkAndConvertListOfFilter(filters []interface{}) (*[]Filter, error) {
	var fs []Filter

	for _, filter := range filters {
		f := filter.(map[string]interface{})
		filterType, ok := f["filterType"]
		if !ok {
			return nil, errors.New("filter doesn't have filterType field")
		}
		filterTypeStr, ok := filterType.(string)
		if !ok {
			return nil, errors.New("filter type is not string")
		}

		filterContent, ok := f["content"]
		if !ok {
			return nil, errors.New("filter doesn't have content field")
		}
		fs = append(fs, Filter{
			FilterType: filterTypeStr,
			Content:    filterContent,
		})
	}

	return &fs, nil
}

const existPodSQL = `EXISTS (SELECT 1 FROM task_pod WHERE task_id=id AND pod_name="%s")`
const existNamespaceSQL = `EXISTS (SELECT 1 FROM task_pod WHERE task_id=id AND pod_namespace="%s")`
