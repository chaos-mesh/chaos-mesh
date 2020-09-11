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

	"github.com/go-logr/logr"
	multierror "github.com/hashicorp/go-multierror"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	v1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dnspb "github.com/chaos-mesh/k8s_dns_chaos/pb"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/common"
	"github.com/chaos-mesh/chaos-mesh/controllers/twophase"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/utils"
)

var (
	// DNSServerName is the chaos DNS server's name
	DNSServerName = "chaos-mesh-dns-service"

	// DNSServerSelectorLabels is the labels used to select the DNS server
	DNSServerSelectorLabels = map[string]string{
		"app.kubernetes.io/component": "dns-server",
		"app.kubernetes.io/instance":  "chaos-mesh",
	}
)

type Reconciler struct {
	client.Client
	client.Reader
	record.EventRecorder
	Log logr.Logger
}

// Reconcile reconciles a DNSChaos resource
func (r *Reconciler) Reconcile(req ctrl.Request, chaos *v1alpha1.DNSChaos) (ctrl.Result, error) {
	r.Log.Info("Reconciling dnschaos")

	scheduler := chaos.GetScheduler()
	duration, err := chaos.GetDuration()
	if err != nil {
		r.Log.Error(err, fmt.Sprintf("unable to get dnschaos[%s/%s]'s duration", chaos.Namespace, chaos.Name))
		return ctrl.Result{}, err
	}
	if scheduler == nil && duration == nil {
		return r.commonDNSChaos(chaos, req)
	} else if scheduler != nil && duration != nil {
		return r.scheduleDNSChaos(chaos, req)
	}

	err = fmt.Errorf("dnschaos[%s/%s] spec invalid", chaos.Namespace, chaos.Name)
	// This should be ensured by admission webhook in the future
	r.Log.Error(err, "scheduler and duration should be omitted or defined at the same time")

	return ctrl.Result{}, nil
}

func (r *Reconciler) commonDNSChaos(dnschaos *v1alpha1.DNSChaos, req ctrl.Request) (ctrl.Result, error) {
	cr := common.NewReconciler(r, r.Client, r.Reader, r.Log)
	return cr.Reconcile(req)
}

func (r *Reconciler) scheduleDNSChaos(dnschaos *v1alpha1.DNSChaos, req ctrl.Request) (ctrl.Result, error) {
	sr := twophase.NewReconciler(r, r.Client, r.Reader, r.Log)
	return sr.Reconcile(req)
}

// Apply applies dns-chaos
func (r *Reconciler) Apply(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	dnschaos, ok := chaos.(*v1alpha1.DNSChaos)
	if !ok {
		err := errors.New("chaos is not dnschaos")
		r.Log.Error(err, "chaos is not DNSChaos", "chaos", chaos)
		return err
	}

	pods, err := utils.SelectAndFilterPods(ctx, r.Client, r.Reader, &dnschaos.Spec)
	if err != nil {
		r.Log.Error(err, "failed to select and generate pods")
		return err
	}

	// get dns server's ip used for chaos
	services, err := utils.SelectAndFilterService(ctx, r.Client, DNSServerSelectorLabels)
	if err != nil {
		r.Log.Error(err, "fail to select service", "labels", DNSServerSelectorLabels)
		return err
	}
	if len(services.Items) != 1 {
		err = fmt.Errorf("get %d services from the labels %v", len(services.Items), DNSServerSelectorLabels)
		r.Log.Error(err, "fail to select service")
		return err
	}
	service := services.Items[0]

	// TODO: get port from dns service to instead 9288
	err = r.setDNSServerRules(service.Spec.ClusterIP, 9288, dnschaos.Name, pods, dnschaos.Spec.Action, dnschaos.Spec.Scope)
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
	r.Event(dnschaos, v1.EventTypeNormal, utils.EventChaosInjected, "")
	return nil
}

// Recover means the reconciler recovers the chaos action
func (r *Reconciler) Recover(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	dnschaos, ok := chaos.(*v1alpha1.DNSChaos)
	if !ok {
		err := errors.New("chaos is not DNSChaos")
		r.Log.Error(err, "chaos is not DNSChaos", "chaos", chaos)
		return err
	}

	// get dns server's ip used for chaos
	services, err := utils.SelectAndFilterService(ctx, r.Client, DNSServerSelectorLabels)
	if err != nil {
		r.Log.Error(err, "fail to select service", "labels", DNSServerSelectorLabels)
		return err
	}
	if len(services.Items) != 1 {
		err = fmt.Errorf("get %d services from the labels %v", len(services.Items), DNSServerSelectorLabels)
		r.Log.Error(err, "fail to select service")
		return err
	}
	service := services.Items[0]
	r.Log.Info("Recover get dns service", "service", service.String(), "ip", service.Spec.ClusterIP)

	// TODO: get port from dns service to instead 9288
	r.cancelDNSServerRules(service.Spec.ClusterIP, 9288, dnschaos.Name)

	if err := r.cleanFinalizersAndRecover(ctx, dnschaos); err != nil {
		return err
	}
	r.Event(dnschaos, v1.EventTypeNormal, utils.EventChaosRecovered, "")

	return nil
}

func (r *Reconciler) cleanFinalizersAndRecover(ctx context.Context, chaos *v1alpha1.DNSChaos) error {
	var result error

	for _, key := range chaos.Finalizers {
		ns, name, err := cache.SplitMetaNamespaceKey(key)
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}

		var pod v1.Pod
		err = r.Client.Get(ctx, types.NamespacedName{
			Namespace: ns,
			Name:      name,
		}, &pod)

		if err != nil {
			if !k8serror.IsNotFound(err) {
				result = multierror.Append(result, err)
				continue
			}

			r.Log.Info("Pod not found", "namespace", ns, "name", name)
			chaos.Finalizers = utils.RemoveFromFinalizer(chaos.Finalizers, key)
			continue
		}

		err = r.recoverPod(ctx, &pod, chaos)
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}

		chaos.Finalizers = utils.RemoveFromFinalizer(chaos.Finalizers, key)
	}

	if chaos.Annotations[common.AnnotationCleanFinalizer] == common.AnnotationCleanFinalizerForced {
		r.Log.Info("Force cleanup all finalizers", "chaos", chaos)
		chaos.Finalizers = chaos.Finalizers[:0]
		return nil
	}

	return result
}

func (r *Reconciler) recoverPod(ctx context.Context, pod *v1.Pod, chaos *v1alpha1.DNSChaos) error {
	r.Log.Info("Try to recover pod", "namespace", pod.Namespace, "name", pod.Name)

	daemonClient, err := utils.NewChaosDaemonClient(ctx, r.Client,
		pod, common.ControllerCfg.ChaosDaemonPort)
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
	})
	if err != nil {
		r.Log.Error(err, "recover pod for DNS chaos")
		return err
	}

	return nil
}

// Object would return the instance of chaos
func (r *Reconciler) Object() v1alpha1.InnerObject {
	return &v1alpha1.DNSChaos{}
}

func (r *Reconciler) applyAllPods(ctx context.Context, pods []v1.Pod, chaos *v1alpha1.DNSChaos, dnsServerIP string) error {
	g := errgroup.Group{}
	for index := range pods {
		pod := &pods[index]

		key, err := cache.MetaNamespaceKeyFunc(pod)
		if err != nil {
			r.Log.Error(err, "get meta namespace key")
			return err
		}
		chaos.Finalizers = utils.InsertFinalizer(chaos.Finalizers, key)

		g.Go(func() error {
			return r.applyPod(ctx, pod, chaos, dnsServerIP)
		})
	}
	err := g.Wait()
	if err != nil {
		r.Log.Error(err, "g.Wait")
		return err
	}
	return nil
}

func (r *Reconciler) applyPod(ctx context.Context, pod *v1.Pod, chaos *v1alpha1.DNSChaos, dnsServerIP string) error {
	r.Log.Info("Try to apply dns chaos", "namespace",
		pod.Namespace, "name", pod.Name)
	daemonClient, err := utils.NewChaosDaemonClient(ctx, r.Client,
		pod, common.ControllerCfg.ChaosDaemonPort)
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
	})
	if err != nil {
		r.Log.Error(err, "set dns server")
		return err
	}

	return nil
}

func (r *Reconciler) setDNSServerRules(dnsServerIP string, port int64, name string, pods []v1.Pod, action v1alpha1.DNSChaosAction, scope v1alpha1.DNSChaosScope) error {
	r.Log.Info("setDNSServerRules")
	defer r.Log.Info("setDNSServerRules finished")
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
		Name:   name,
		Action: string(action),
		Pods:   pbPods,
		Scope:  string(scope),
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

func (r *Reconciler) cancelDNSServerRules(dnsServerIP string, port int64, name string) error {
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
