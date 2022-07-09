package chaosdaemon

import (
	"golang.org/x/net/context"
	"net"
	"net/http"
)

type unixSocketTransport struct {
	addr string
}

func (t unixSocketTransport) dial(ctx context.Context, network, addr string) (net.Conn, error) {
	return net.Dial("unix", t.addr)
}

func (t *unixSocketTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	client := http.Client{Transport: &http.Transport{DialContext: t.dial}}
	resp, err = client.Do(req)

	return
}
