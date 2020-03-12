// Copyright 2019 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package portforward

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"k8s.io/klog"
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
	ctx    context.Context
	config *rest.Config
	client kubernetes.Interface
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
			klog.Infof("log from port forwarding %q: %s", podKey, lineScanner.Text())
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

	pod, err := f.client.CoreV1().Pods(namespace).Get(forwardablePod.Name, metav1.GetOptions{})
	if err != nil {
		return nil, nil, err
	}
	return f.ForwardPod(pod, addresses, ports)
}

func (f *portForwarder) ForwardPod(pod *corev1.Pod, addresses []string, ports []string) (forwardedPorts []portforward.ForwardedPort, cancel context.CancelFunc, err error) {
	if pod.Status.Phase != corev1.PodRunning {
		return nil, nil, fmt.Errorf("unable to forward port because pod is not running. Current status=%v", pod.Status.Phase)
	}

	req := f.client.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(pod.Namespace).
		Name(pod.Name).
		SubResource("portforward")

	return f.forwardPorts(fmt.Sprintf("%s/%s", pod.Namespace, pod.Name), "POST", req.URL(), addresses, ports)
}

func NewPortForwarder(ctx context.Context, restClientGetter genericclioptions.RESTClientGetter) (PortForward, error) {
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
	}
	return f, nil
}
