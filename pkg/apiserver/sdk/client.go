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

func (c *Client) CreateJob(job *types.Job) error {
	return c.httpRequest("PUT", "/job", job, nil)
}

func (c *Client) GetJob(filter []*filter.Filter) ([]types.Job, error) {
	encodedFilter, err := json.Marshal(filter)
	if err != nil {
		log.Error(err)
		return nil, errors.Trace(err)
	}

	queryFilter := url.QueryEscape(string(encodedFilter))

	var jobs []types.Job
	err = c.httpRequest("GET", fmt.Sprintf("/jobs?filters=%s", queryFilter), nil, &jobs)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return jobs, nil
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
