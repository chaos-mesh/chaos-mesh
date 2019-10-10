package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/juju/errors"
	"github.com/ngaut/log"
	"github.com/pingcap/chaos-operator/pkg/apiserver"
	"github.com/pingcap/chaos-operator/pkg/apiserver/filter"
	"github.com/pingcap/chaos-operator/pkg/apiserver/types"
	v1 "k8s.io/api/core/v1"
)

type Client struct {
	address string
}

func (c *Client) httpRequest(method string, path string, body interface{}, resp interface{}) error {
	client := &http.Client{}

	httpBody, err := json.Marshal(body)
	if err != nil {
		log.Error(err)
		return errors.Trace(err)
	}

	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", c.address, path), bytes.NewReader(httpBody))
	if err != nil {
		log.Error(err)
		return errors.Trace(err)
	}

	res, err := client.Do(req)
	if err != nil {
		log.Error(err)
		return errors.Trace(err)
	}

	var response apiserver.Response
	if resp != nil {
		response.Data = resp
	}

	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&response)
	if err != nil {
		log.Error(err)
		return errors.Trace(err)
	}

	if response.Code != apiserver.StatusOK {
		return errors.New(fmt.Sprintf("error %d: %s", response.Code, response.Message))
	}

	return nil
}

func (c *Client) CreateTask(task *types.Task) error {
	return c.httpRequest("PUT", "/task", task, nil)
}

func (c *Client) CreatePodsTask(pods []v1.Pod, taskType string, eventType string, resource interface{}) error {
	podList := []types.Pod{}

	for _, pod := range pods {
		podList = append(podList, types.Pod{
			Name:      pod.GetName(),
			Namespace: pod.GetNamespace(),
		})
	}

	return c.CreateTask(&types.Task{
		Pods:      podList,
		TaskType:  taskType,
		EventType: eventType,
		Resource:  resource,
	})
}

func (c *Client) GetTask(filter *filter.Filter) ([]types.Task, error) {
	encodedFilter, err := json.Marshal(filter)
	if err != nil {
		log.Error(err)
		return nil, errors.Trace(err)
	}

	queryFilter := url.QueryEscape(string(encodedFilter))

	var tasks []types.Task
	err = c.httpRequest("GET", fmt.Sprintf("/tasks?filters=%s", queryFilter), nil, &tasks)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return tasks, nil
}

// NewClientFromEnv will create a client with discovering services
func NewClientInK8s() *Client {
	clientHost := os.Getenv("CHAOS_API_SERVER_SERVICE_HOST")
	clientPort := os.Getenv("CHAOS_API_SERVER_SERVICE_PORT")

	// TODO: check error here
	client := Client{
		address: fmt.Sprintf("http://%s:%s", clientHost, clientPort),
	}

	return &client
}
