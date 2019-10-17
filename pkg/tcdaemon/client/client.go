package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/golang/glog"
	"github.com/juju/errors"
	"github.com/ngaut/log"
	"github.com/pingcap/chaos-operator/pkg/tcdaemon"
)

type Client struct {
	host string
	port string
}

func (c *Client) httpRequest(method string, path string, query map[string]string, body interface{}, resp interface{}) error {
	client := &http.Client{}

	httpBody, err := json.Marshal(body)
	if err != nil {
		log.Error(err)
		return errors.Trace(err)
	}

	u := url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%s", c.host, c.port),
		Path:   path,
	}
	q := u.Query()
	for key, value := range query {
		q.Add(key, value)
	}
	u.RawQuery = q.Encode()

	glog.Infof("sending request to %s", u.String())
	req, err := http.NewRequest(method, u.String(), bytes.NewReader(httpBody))
	if err != nil {
		log.Error(err)
		return errors.Trace(err)
	}

	res, err := client.Do(req)
	if err != nil {
		log.Error(err)
		return errors.Trace(err)
	}

	var response tcdaemon.Response
	if resp != nil {
		response.Data = resp
	}

	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&response)
	if err != nil {
		log.Error(err)
		return errors.Trace(err)
	}

	if response.Code != tcdaemon.StatusOK {
		return errors.New(fmt.Sprintf("error %d: %s", response.Code, response.Message))
	}

	return nil
}

func (c *Client) AddNetem(containerID string, netem *tcdaemon.Netem) error {
	return c.httpRequest("PUT", "/netem", map[string]string{
		"containerID": containerID,
	}, netem, nil)
}

func (c *Client) DeleteNetem(containerID string) error {
	return c.httpRequest("DELETE", "/netem", map[string]string{
		"containerID": containerID,
	}, nil, nil)
}

func NewClient(host string, port string) *Client {
	return &Client{
		host,
		port,
	}
}
