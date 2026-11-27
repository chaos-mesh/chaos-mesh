// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package portforward

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/apimachinery/pkg/util/httpstream/spdy"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	forward "k8s.io/client-go/tools/portforward"
)

type forwardConfig struct {
	genericclioptions.RESTClientGetter
	config  *rest.Config
	failure error
}

func (f forwardConfig) ToRESTConfig() (*rest.Config, error) {
	return f.config, f.failure
}

func (forwardConfig) ToRESTMapper() (meta.RESTMapper, error) {
	mapper := meta.NewDefaultRESTMapper([]schema.GroupVersion{corev1.SchemeGroupVersion})
	mapper.Add(corev1.SchemeGroupVersion.WithKind("Pod"), meta.RESTScopeNamespace)
	return mapper, nil
}

func TestForwardOnePortRelaysDataFromNamedPod(t *testing.T) {
	g := NewWithT(t)
	const podPath = "/api/v1/namespaces/work/pods/echo"
	upgraded := make(chan struct{}, 1)
	disconnected := make(chan struct{}, 1)
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch {
		case req.Method == http.MethodGet && req.URL.Path == podPath:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(&corev1.Pod{
				TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Pod"},
				ObjectMeta: metav1.ObjectMeta{Namespace: "work", Name: "echo"},
				Status:     corev1.PodStatus{Phase: corev1.PodRunning},
			})
		case req.Method == http.MethodPost && req.URL.Path == podPath+"/portforward":
			_, err := httpstream.Handshake(req, w, []string{forward.PortForwardProtocolV1Name})
			if err != nil {
				return
			}
			conn := spdy.NewResponseUpgrader().UpgradeResponse(w, req, handleEchoStream)
			if conn != nil {
				upgraded <- struct{}{}
				<-conn.CloseChan()
				disconnected <- struct{}{}
			}
		default:
			http.NotFound(w, req)
		}
	}))
	t.Cleanup(api.Close)

	ctx, stop := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(stop)
	client, err := NewPortForwarder(ctx, forwardConfig{config: &rest.Config{Host: api.URL, Timeout: 5 * time.Second}}, true, logr.Discard())
	g.Expect(err).NotTo(HaveOccurred())
	address, port, cancel, err := ForwardOnePort(client, "work", "echo", 8080)
	g.Expect(err).NotTo(HaveOccurred())
	t.Cleanup(cancel)
	g.Expect(address).To(Equal("127.0.0.1"))
	g.Expect(port).NotTo(BeZero())
	g.Eventually(upgraded).WithTimeout(2 * time.Second).Should(Receive())

	localAddress := net.JoinHostPort(address, strconv.Itoa(int(port)))
	conn, err := net.DialTimeout("tcp", localAddress, time.Second)
	g.Expect(err).NotTo(HaveOccurred())
	t.Cleanup(func() { _ = conn.Close() })
	g.Expect(conn.SetDeadline(time.Now().Add(time.Second))).To(Succeed())
	_, err = io.WriteString(conn, "ping")
	g.Expect(err).NotTo(HaveOccurred())
	reply := make([]byte, len("pong"))
	_, err = io.ReadFull(conn, reply)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(reply)).To(Equal("pong"))

	cancel()
	g.Eventually(disconnected).WithTimeout(2 * time.Second).Should(Receive())
	g.Eventually(func() error {
		reopened, err := net.DialTimeout("tcp", localAddress, 100*time.Millisecond)
		if reopened != nil {
			_ = reopened.Close()
		}
		return err
	}).WithTimeout(2 * time.Second).Should(HaveOccurred())
}

func handleEchoStream(stream httpstream.Stream, ready <-chan struct{}) error {
	if stream.Headers().Get(corev1.PortHeader) != "8080" {
		return errors.New("unexpected remote port")
	}
	go func() {
		<-ready
		defer stream.Close()
		if stream.Headers().Get(corev1.StreamType) != corev1.StreamTypeData {
			return
		}
		input := make([]byte, len("ping"))
		if _, err := io.ReadFull(stream, input); err == nil && string(input) == "ping" {
			_, _ = io.WriteString(stream, "pong")
		}
	}()
	return nil
}

func TestForwardPodRejectsPendingPodsAndFailedUpgrades(t *testing.T) {
	g := NewWithT(t)
	failure := errors.New("cannot load Kubernetes config")
	g.Expect(NewPortForwarder(t.Context(), forwardConfig{failure: failure}, false, logr.Discard())).Error().To(MatchError(failure))
	g.Expect(NewPortForwarder(t.Context(), forwardConfig{config: &rest.Config{Host: "://invalid"}}, false, logr.Discard())).Error().To(HaveOccurred())

	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "port forwarding disabled", http.StatusForbidden)
	}))
	t.Cleanup(api.Close)
	client, err := NewPortForwarder(t.Context(), forwardConfig{config: &rest.Config{Host: api.URL}}, false, logr.Discard())
	g.Expect(err).NotTo(HaveOccurred())
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "work", Name: "echo"}}
	pod.Status.Phase = corev1.PodPending
	_, _, err = client.ForwardPod(pod, []string{"127.0.0.1"}, []string{"0:8080"})
	g.Expect(err).To(MatchError(ContainSubstring("pod is not running")))
	pod.Status.Phase = corev1.PodRunning
	_, _, err = client.ForwardPod(pod, []string{"127.0.0.1"}, []string{"invalid"})
	g.Expect(err).To(HaveOccurred())
	_, _, err = client.ForwardPod(pod, []string{"127.0.0.1"}, []string{"0:8080"})
	g.Expect(err).To(MatchError(ContainSubstring("port forwarding disabled")))
}
