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

package dnschaos

import (
	"context"
	"fmt"
	"net"
	"time"

	dnspb "github.com/chaos-mesh/k8s_dns_chaos/pb"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	impltypes "github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/types"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/utils"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

var _ impltypes.ChaosImpl = (*Impl)(nil)

type Impl struct {
	client.Client
	Log logr.Logger

	decoder *utils.ContainerRecordDecoder
}

func (impl *Impl) Apply(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	decodedContainer, err := impl.decoder.DecodeContainerRecord(ctx, records[index], obj)
	if decodedContainer.PbClient != nil {
		defer decodedContainer.PbClient.Close()
	}
	if err != nil {
		return v1alpha1.NotInjected, err
	}

	service, err := impl.getService(ctx, config.ControllerCfg.Namespace, config.ControllerCfg.DNSServiceName)
	if err != nil {
		impl.Log.Error(err, "fail to get dns service")
		return v1alpha1.NotInjected, err
	}

	dnsPods, err := impl.getPodsFromSelector(ctx, config.ControllerCfg.Namespace, service.Spec.Selector)
	if err != nil {
		impl.Log.Error(err, "fail to get pods from selector")
		return v1alpha1.NotInjected, err
	}

	dnschaos := obj.(*v1alpha1.DNSChaos)
	for _, pod := range dnsPods {
		err = impl.setDNSServerRules(pod.Status.PodIP, config.ControllerCfg.DNSServicePort, dnschaos.Name, decodedContainer.Pod, dnschaos.Spec.Action, dnschaos.Spec.DomainNamePatterns)
		if err != nil {
			impl.Log.Error(err, "fail to set DNS server rules")
			return v1alpha1.NotInjected, err
		}
		impl.Log.Info("Apply DNS chaos to DNS pod", "ip", service.Spec.ClusterIP)
	}

	_, err = decodedContainer.PbClient.SetDNSServer(ctx, &pb.SetDNSServerRequest{
		ContainerId: decodedContainer.ContainerId,
		DnsServer:   service.Spec.ClusterIP,
		Enable:      true,
		EnterNS:     true,
	})
	if err != nil {
		impl.Log.Error(err, "set dns server")
		return v1alpha1.NotInjected, err
	}

	return v1alpha1.Injected, nil
}

func (impl *Impl) setDNSServerRules(dnsServerIP string, port int, name string, pod *v1.Pod, action v1alpha1.DNSChaosAction, patterns []string) error {
	impl.Log.Info("setDNSServerRules", "name", name)

	pbPods := make([]*dnspb.Pod, 1)
	pbPods[0] = &dnspb.Pod{
		Name:      pod.Name,
		Namespace: pod.Namespace,
	}

	conn, err := grpc.Dial(net.JoinHostPort(dnsServerIP, fmt.Sprintf("%d", port)), grpc.WithInsecure())
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
		return errors.Errorf("set dns chaos to dns server error: %s", response.Msg)
	}

	return nil
}

func (impl *Impl) Recover(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	decodedContainer, err := impl.decoder.DecodeContainerRecord(ctx, records[index], obj)
	if decodedContainer.PbClient != nil {
		defer decodedContainer.PbClient.Close()
	}
	if err != nil {
		if errors.Is(err, utils.ErrContainerNotFound) {
			// pretend the disappeared container has been recovered
			return v1alpha1.NotInjected, nil
		}
		return v1alpha1.Injected, err
	}

	dnschaos := obj.(*v1alpha1.DNSChaos)

	// get dns server's ip used for chaos
	service, err := impl.getService(ctx, config.ControllerCfg.Namespace, config.ControllerCfg.DNSServiceName)
	if err != nil {
		impl.Log.Error(err, "fail to get dns service")
		return v1alpha1.Injected, err
	}

	dnsPods, err := impl.getPodsFromSelector(ctx, config.ControllerCfg.Namespace, service.Spec.Selector)
	if err != nil {
		impl.Log.Error(err, "fail to get pods from selector")
		return v1alpha1.NotInjected, err
	}

	for _, pod := range dnsPods {
		err = impl.cancelDNSServerRules(pod.Status.PodIP, config.ControllerCfg.DNSServicePort, dnschaos.Name)
		if err != nil {
			impl.Log.Error(err, "fail to cancelDNSServerRules")
			return v1alpha1.Injected, err
		}
		impl.Log.Info("Cancel DNS chaos to DNS pod", "ip", service.Spec.ClusterIP)
	}

	_, err = decodedContainer.PbClient.SetDNSServer(ctx, &pb.SetDNSServerRequest{
		ContainerId: decodedContainer.ContainerId,
		Enable:      false,
		EnterNS:     true,
	})
	if err != nil {
		impl.Log.Error(err, "recover pod for DNS chaos")
		return v1alpha1.Injected, err
	}

	return v1alpha1.NotInjected, err
}

func (impl *Impl) cancelDNSServerRules(dnsServerIP string, port int, name string) error {
	conn, err := grpc.Dial(net.JoinHostPort(dnsServerIP, fmt.Sprintf("%d", port)), grpc.WithInsecure())
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
		return errors.Errorf("set dns chaos to dns server error %s", response.Msg)
	}

	return nil
}

// getService get k8s service by service name
func (impl *Impl) getService(ctx context.Context, namespace string, serviceName string) (*v1.Service, error) {
	service := &v1.Service{}
	err := impl.Client.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      serviceName,
	}, service)
	if err != nil {
		return nil, err
	}

	return service, nil
}

// getPodsFromSelector returns the pods assiocated to a given service
func (impl *Impl) getPodsFromSelector(ctx context.Context, namespace string, labelSelector map[string]string) ([]v1.Pod, error) {
	lSelector := labels.SelectorFromSet(labelSelector)
	listOptions := &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: lSelector,
	}
	podsList := &v1.PodList{}
	err := impl.Client.List(ctx, podsList, listOptions)
	if err != nil {
		return nil, err
	}

	return podsList.Items, nil
}

func NewImpl(c client.Client, log logr.Logger, decoder *utils.ContainerRecordDecoder) *impltypes.ChaosImplPair {
	return &impltypes.ChaosImplPair{
		Name:   "dnschaos",
		Object: &v1alpha1.DNSChaos{},
		Impl: &Impl{
			Client:  c,
			Log:     log.WithName("dnschaos"),
			decoder: decoder,
		},
	}
}

var Module = fx.Provide(
	fx.Annotated{
		Group:  "impl",
		Target: NewImpl,
	},
)
