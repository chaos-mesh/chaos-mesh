// Copyright 2020 Chaos Mesh Authors.
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

package dnschaos

import (
	"context"
	"errors"
	"fmt"
	"time"

	dnspb "github.com/chaos-mesh/k8s_dns_chaos/pb"
	"github.com/go-logr/logr"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
	kubeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/controllers/recover"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/client"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/events"
	"github.com/chaos-mesh/chaos-mesh/pkg/finalizer"
	"github.com/chaos-mesh/chaos-mesh/pkg/router"
	ctx "github.com/chaos-mesh/chaos-mesh/pkg/router/context"
	end "github.com/chaos-mesh/chaos-mesh/pkg/router/endpoint"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector"
)

// endpoint is dns-chaos reconciler
type endpoint struct {
	ctx.Context
}

type recoverer struct {
	kubeclient.Client
	Log logr.Logger
}

// Apply applies dns-chaos
func (r *endpoint) Apply(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	dnschaos, ok := chaos.(*v1alpha1.DNSChaos)
	if !ok {
		err := errors.New("chaos is not DNSChaos")
		r.Log.Error(err, "chaos is not DNSChaos", "chaos", chaos)
		return err
	}

	pods, err := selector.SelectAndFilterPods(ctx, r.Client, r.Reader, &dnschaos.Spec, config.ControllerCfg.ClusterScoped, config.ControllerCfg.TargetNamespace, config.ControllerCfg.AllowedNamespaces, config.ControllerCfg.IgnoredNamespaces)
	if err != nil {
		r.Log.Error(err, "failed to select and generate pods")
		return err
	}

	// get dns server's ip used for chaos
	service, err := selector.GetService(ctx, r.Client, "", config.ControllerCfg.Namespace, config.ControllerCfg.DNSServiceName)
	if err != nil {
		r.Log.Error(err, "fail to get service")
		return err
	}
	r.Log.Info("Set DNS chaos to DNS service", "ip", service.Spec.ClusterIP)

	err = r.setDNSServerRules(service.Spec.ClusterIP, config.ControllerCfg.DNSServicePort, dnschaos.Name, pods, dnschaos.Spec.Action, dnschaos.Spec.DomainNamePatterns)
	if err != nil {
		r.Log.Error(err, "fail to set DNS server rules")
		return err
	}

	if err = r.applyAllPods(ctx, pods, dnschaos, service.Spec.ClusterIP); err != nil {
		r.Log.Error(err, "failed to apply chaos on all pods")
		return err
	}

	dnschaos.Status.Experiment.PodRecords = make([]v1alpha1.PodStatus, 0, len(pods))
	for _, pod := range pods {
		ps := v1alpha1.PodStatus{
			Namespace: pod.Namespace,
			Name:      pod.Name,
			HostIP:    pod.Status.HostIP,
			PodIP:     pod.Status.PodIP,
		}

		dnschaos.Status.Experiment.PodRecords = append(dnschaos.Status.Experiment.PodRecords, ps)
	}
	r.Event(dnschaos, v1.EventTypeNormal, events.ChaosInjected, "")
	return nil
}

// Recover means the reconciler recovers the chaos action
func (r *endpoint) Recover(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	dnschaos, ok := chaos.(*v1alpha1.DNSChaos)
	if !ok {
		err := errors.New("chaos is not DNSChaos")
		r.Log.Error(err, "chaos is not DNSChaos", "chaos", chaos)
		return err
	}

	// get dns server's ip used for chaos
	service, err := selector.GetService(ctx, r.Client, "", config.ControllerCfg.Namespace, config.ControllerCfg.DNSServiceName)
	if err != nil {
		r.Log.Error(err, "fail to get service")
		return err
	}
	r.Log.Info("Cancel DNS chaos to DNS service", "ip", service.Spec.ClusterIP)

	r.cancelDNSServerRules(service.Spec.ClusterIP, config.ControllerCfg.DNSServicePort, dnschaos.Name)

	rd := recover.Delegate{Client: r.Client, Log: r.Log, RecoverIntf: &recoverer{r.Client, r.Log}}

	finalizers, err := rd.CleanFinalizersAndRecover(ctx, chaos, dnschaos.Finalizers, dnschaos.Annotations)
	if err != nil {
		return err
	}
	dnschaos.Finalizers = finalizers
	r.Event(dnschaos, v1.EventTypeNormal, events.ChaosRecovered, "")

	return nil
}

func (r *recoverer) RecoverPod(ctx context.Context, pod *v1.Pod, somechaos v1alpha1.InnerObject) error {
	r.Log.Info("Try to recover pod", "namespace", pod.Namespace, "name", pod.Name)

	daemonClient, err := client.NewChaosDaemonClient(ctx, r.Client,
		pod, config.ControllerCfg.ChaosDaemonPort)
	if err != nil {
		r.Log.Error(err, "get chaos daemon client")
		return err
	}
	defer daemonClient.Close()
	if len(pod.Status.ContainerStatuses) == 0 {
		return fmt.Errorf("%s %s can't get the state of container", pod.Namespace, pod.Name)
	}

	target := pod.Status.ContainerStatuses[0].ContainerID

	_, err = daemonClient.SetDNSServer(ctx, &pb.SetDNSServerRequest{
		ContainerId: target,
		Enable:      false,
		EnterNS:     true,
	})
	if err != nil {
		r.Log.Error(err, "recover pod for DNS chaos")
		return err
	}

	return nil
}

// Object would return the instance of chaos
func (r *endpoint) Object() v1alpha1.InnerObject {
	return &v1alpha1.DNSChaos{}
}

func (r *endpoint) applyAllPods(ctx context.Context, pods []v1.Pod, chaos *v1alpha1.DNSChaos, dnsServerIP string) error {
	g := errgroup.Group{}
	for index := range pods {
		pod := &pods[index]

		key, err := cache.MetaNamespaceKeyFunc(pod)
		if err != nil {
			r.Log.Error(err, "get meta namespace key")
			return err
		}
		chaos.Finalizers = finalizer.InsertFinalizer(chaos.Finalizers, key)

		g.Go(func() error {
			return r.applyPod(ctx, pod, dnsServerIP)
		})
	}
	err := g.Wait()
	if err != nil {
		r.Log.Error(err, "g.Wait")
		return err
	}
	return nil
}

func (r *endpoint) applyPod(ctx context.Context, pod *v1.Pod, dnsServerIP string) error {
	r.Log.Info("Try to apply dns chaos", "namespace",
		pod.Namespace, "name", pod.Name)
	daemonClient, err := client.NewChaosDaemonClient(ctx, r.Client,
		pod, config.ControllerCfg.ChaosDaemonPort)
	if err != nil {
		r.Log.Error(err, "get chaos daemon client")
		return err
	}
	defer daemonClient.Close()
	if len(pod.Status.ContainerStatuses) == 0 {
		return fmt.Errorf("%s %s can't get the state of container", pod.Namespace, pod.Name)
	}

	target := pod.Status.ContainerStatuses[0].ContainerID

	_, err = daemonClient.SetDNSServer(ctx, &pb.SetDNSServerRequest{
		ContainerId: target,
		DnsServer:   dnsServerIP,
		Enable:      true,
		EnterNS:     true,
	})
	if err != nil {
		r.Log.Error(err, "set dns server")
		return err
	}

	return nil
}

func (r *endpoint) setDNSServerRules(dnsServerIP string, port int, name string, pods []v1.Pod, action v1alpha1.DNSChaosAction, patterns []string) error {
	r.Log.Info("setDNSServerRules", "name", name)

	pbPods := make([]*dnspb.Pod, len(pods))
	for i, pod := range pods {
		pbPods[i] = &dnspb.Pod{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		}
	}

	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", dnsServerIP, port), grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	c := dnspb.NewDNSClient(conn)
	request := &dnspb.SetDNSChaosRequest{
		Name:     name,
		Action:   string(action),
		Pods:     pbPods,
		Patterns: patterns,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	response, err := c.SetDNSChaos(ctx, request)
	if err != nil {
		return err
	}

	if !response.Result {
		return fmt.Errorf("set dns chaos to dns server error %s", response.Msg)
	}

	return nil
}

func (r *endpoint) cancelDNSServerRules(dnsServerIP string, port int, name string) error {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", dnsServerIP, port), grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	c := dnspb.NewDNSClient(conn)
	request := &dnspb.CancelDNSChaosRequest{
		Name: name,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	response, err := c.CancelDNSChaos(ctx, request)
	if err != nil {
		return err
	}

	if !response.Result {
		return fmt.Errorf("set dns chaos to dns server error %s", response.Msg)
	}

	return nil
}

func init() {
	router.Register("dnschaos", &v1alpha1.DNSChaos{}, func(obj runtime.Object) bool {
		return true
	}, func(ctx ctx.Context) end.Endpoint {
		return &endpoint{
			Context: ctx,
		}
	})
}
