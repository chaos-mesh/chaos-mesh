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
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"k8s.io/kubectl/pkg/polymorphichelpers"
)

const (
	getPodTimeout = time.Minute
)

// PortForward represents an interface which can forward local ports to a pod.
type PortForward interface {
	Forward(namespace, resourceName string, addresses []string, ports []string) ([]portforward.ForwardedPort, context.CancelFunc, error)
	ForwardPod(pod *corev1.Pod, addresses []string, ports []string) ([]portforward.ForwardedPort, context.CancelFunc, error)
}

// portForwarder implements PortForward interface
type portForwarder struct {
	genericclioptions.RESTClientGetter
	ctx       context.Context
	config    *rest.Config
	client    kubernetes.Interface
	enableLog bool
	logger    logr.Logger
}

var _ PortForward = &portForwarder{}

func (f *portForwarder) forwardPorts(podKey, method string, url *url.URL, addresses []string, ports []string) (forwardedPorts []portforward.ForwardedPort, cancel context.CancelFunc, err error) {
	transport, upgrader, err := spdy.RoundTripperFor(f.config)
	if err != nil {
		return nil, nil, err
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, method, url)
	r, w := io.Pipe()
	ctx, cancel := context.WithCancel(f.ctx)
	readyChan := make(chan struct{})
	fw, err := portforward.NewOnAddresses(dialer, addresses, ports, ctx.Done(), readyChan, w, w)
	if err != nil {
		return nil, nil, err
	}

	// logging stdout/stderr of port forwarding
	go func() {
		// close pipe if the context is done
		<-ctx.Done()
		w.Close()
	}()

	go func() {
		lineScanner := bufio.NewScanner(r)
		for lineScanner.Scan() {
			if f.enableLog {
				f.logger.Info(fmt.Sprintf("log from port forwarding %q: %s", podKey, lineScanner.Text()))
			}
		}
	}()

	// run port forwarding
	errChan := make(chan error)
	go func() {
		errChan <- fw.ForwardPorts()
	}()

	// wait for ready or error
	select {
	case <-readyChan:
		break
	case err := <-errChan:
		cancel()
		return nil, nil, err
	}

	forwardedPorts, err = fw.GetPorts()
	if err != nil {
		cancel()
		return nil, nil, err
	}

	return forwardedPorts, cancel, nil
}

// Forward would port-forward to target resources
func (f *portForwarder) Forward(namespace, resourceName string, addresses []string, ports []string) (forwardedPorts []portforward.ForwardedPort, cancel context.CancelFunc, err error) {
	builder := resource.NewBuilder(f).
		WithScheme(scheme.Scheme, scheme.Scheme.PrioritizedVersionsAllGroups()...).
		ContinueOnError().
		NamespaceParam(namespace).DefaultNamespace()

	builder.ResourceNames("pods", resourceName)

	obj, err := builder.Do().Object()
	if err != nil {
		return nil, nil, err
	}

	forwardablePod, err := polymorphichelpers.AttachablePodForObjectFn(f, obj, getPodTimeout)
	if err != nil {
		return nil, nil, err
	}

	// FIXME: get context from parameter
	pod, err := f.client.CoreV1().Pods(namespace).Get(context.TODO(), forwardablePod.Name, metav1.GetOptions{})
	if err != nil {
		return nil, nil, err
	}
	return f.ForwardPod(pod, addresses, ports)
}

// ForwardPod would port-forward target Pod
func (f *portForwarder) ForwardPod(pod *corev1.Pod, addresses []string, ports []string) (forwardedPorts []portforward.ForwardedPort, cancel context.CancelFunc, err error) {
	if pod.Status.Phase != corev1.PodRunning {
		return nil, nil, errors.Errorf("unable to forward port because pod is not running. Current status=%v", pod.Status.Phase)
	}

	req := f.client.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(pod.Namespace).
		Name(pod.Name).
		SubResource("portforward")

	return f.forwardPorts(fmt.Sprintf("%s/%s", pod.Namespace, pod.Name), "POST", req.URL(), addresses, ports)
}

// NewPortForwarder would create a new port-forward
func NewPortForwarder(ctx context.Context, restClientGetter genericclioptions.RESTClientGetter, enableLog bool, logger logr.Logger) (PortForward, error) {
	config, err := restClientGetter.ToRESTConfig()
	if err != nil {
		return nil, err
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	f := &portForwarder{
		RESTClientGetter: restClientGetter,
		ctx:              ctx,
		config:           config,
		client:           client,
		enableLog:        enableLog,
		logger:           logger,
	}
	return f, nil
}

// ForwardOnePort help to utility to forward one port of Kubernetes resource.
func ForwardOnePort(fw PortForward, ns, resource string, port uint16) (string, uint16, context.CancelFunc, error) {
	ports := []string{fmt.Sprintf("0:%d", port)}
	forwardedPorts, cancel, err := fw.Forward(ns, resource, []string{"127.0.0.1"}, ports)
	if err != nil {
		return "", 0, nil, err
	}
	var localPort uint16
	var found bool
	for _, p := range forwardedPorts {
		if p.Remote == port {
			localPort = p.Local
			found = true
		}
	}
	if !found {
		cancel()
		return "", 0, nil, errors.New("unexpected error")
	}
	return "127.0.0.1", localPort, cancel, nil
}
