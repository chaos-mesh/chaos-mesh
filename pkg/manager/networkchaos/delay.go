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

package networkchaos

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/golang/glog"
	"github.com/pingcap/chaos-operator/pkg/apis/pingcap.com/v1alpha1"
	"github.com/pingcap/chaos-operator/pkg/manager"
	pb "github.com/pingcap/chaos-operator/pkg/tcdaemon/pb"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
)

// DelayJob represents a job to add delay on pod
type DelayJob struct {
	networkChaos *v1alpha1.NetworkChaos
	kubeCli      kubernetes.Interface
	podLister    corelisters.PodLister
}

// Run is the core logic to execute delay chaos experiment.
func (d *DelayJob) Run() {
	var err error

	pods, err := manager.SelectPods(d.networkChaos.Spec.Selector, d.podLister, d.kubeCli)
	if err != nil {
		glog.Errorf("%s, fail to get selected pods, %v", d.logPrefix(), err)
		return
	}

	if pods == nil || len(pods) == 0 {
		glog.Errorf("%s, no pod is selected", d.logPrefix())
		return
	}

	glog.Infof("%s, Try to delay pod network", d.logPrefix())

	duration, err := time.ParseDuration(d.networkChaos.Spec.Duration)
	if err != nil {
		glog.Errorf("fail to parse duration %v", err)
	}
	g := errgroup.Group{}
	for _, pod := range pods {
		pod := pod
		g.Go(func() error {
			err := d.DelayPod(pod)
			if err != nil {
				glog.Errorf("error while delaying pod %v", err)
			}

			time.Sleep(duration)

			return d.ResumePod(pod)
		})
	}

	err = g.Wait()
	if err != nil {
		glog.Errorf("%s, fail to run action, %v", d.logPrefix(), err)
	}
}

func (d *DelayJob) createGrpcConnection(pod v1.Pod) (*grpc.ClientConn, error) {
	port := os.Getenv("TC_DAEMON_PORT")
	if port == "" {
		port = "8080"
	}

	nodeName := pod.Spec.NodeName
	glog.Infof("%s, Creating client to tcdaemon on %s", d.logPrefix(), nodeName)
	node, err := d.kubeCli.CoreV1().Nodes().Get(nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", node.Status.Addresses[0].Address, port), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// DelayPod will add a netem on container network interface to delay it
func (d *DelayJob) DelayPod(pod v1.Pod) error {
	delay := d.networkChaos.Spec.Delay

	glog.Infof("%s, Try to delay pod %s/%s", d.logPrefix(), pod.Namespace, pod.Name)

	c, err := d.createGrpcConnection(pod)
	if err != nil {
		return err
	}
	defer c.Close()

	client := pb.NewTcDaemonClient(c)

	containerId := pod.Status.ContainerStatuses[0].ContainerID

	delayTime, err := time.ParseDuration(delay.Latency)
	if err != nil {
		glog.Errorf("fail to parse delay time %v", err)
	}
	jitter, err := time.ParseDuration(delay.Jitter)
	if err != nil {
		glog.Errorf("fail to parse delay jitter %v", err)
	}

	_, err = client.SetNetem(context.Background(), &pb.NetemRequest{
		ContainerId: containerId,
		Netem: &pb.Netem{
			Time:      uint32(delayTime.Nanoseconds() / 1e3),
			DelayCorr: delay.Correlation,
			Jitter:    uint32(jitter.Nanoseconds() / 1e3),
		},
	})

	return err
}

// Resume will remove every netem from Pod
func (d *DelayJob) ResumePod(pod v1.Pod) error {
	glog.Infof("%s, Try to resume pod %s/%s", d.logPrefix(), pod.Namespace, pod.Name)

	c, err := d.createGrpcConnection(pod)
	if err != nil {
		return err
	}
	defer c.Close()

	client := pb.NewTcDaemonClient(c)

	containerId := pod.Status.ContainerStatuses[0].ContainerID
	_, err = client.DeleteNetem(context.Background(), &pb.NetemRequest{
		ContainerId: containerId,
		Netem:       nil,
	})

	return err
}

// Equal returns true when the two jobs have same NetworkChaos.
// It can be used to judge if the job need to update this job.
func (d *DelayJob) Equal(job manager.Job) bool {
	djob, ok := job.(*DelayJob)
	if !ok {
		return false
	}

	if d.networkChaos.Name != djob.networkChaos.Name ||
		d.networkChaos.Namespace != djob.networkChaos.Namespace {
		return false
	}

	// judge ResourceVersion,
	// If them are same, we can think that the NetworkChaos resource has not changed.
	if d.networkChaos.ResourceVersion != djob.networkChaos.ResourceVersion {
		return false
	}

	return true
}

// Close stops delay job, we need to clean up running delay job (by removing every netem)
func (d *DelayJob) Close() error {
	var err error

	pods, err := manager.SelectPods(d.networkChaos.Spec.Selector, d.podLister, d.kubeCli)
	if err != nil {
		glog.Errorf("%s, fail to get selected pods, %v", d.logPrefix(), err)
		return err
	}

	if pods == nil || len(pods) == 0 {
		glog.Errorf("%s, no pod is selected", d.logPrefix())
		return err
	}

	glog.Infof("%s, Try to resume pod network", d.logPrefix())

	g := errgroup.Group{}
	for _, pod := range pods {
		pod := pod
		g.Go(func() error {
			return d.ResumePod(pod)
		})
	}

	err = g.Wait()
	if err != nil {
		glog.Errorf("%s, fail to run action, %v", d.logPrefix(), err)
		return err
	}

	return nil
}

// Clean does nothing
func (d *DelayJob) Clean() error {
	return nil
}

func (d *DelayJob) logPrefix() string {
	return fmt.Sprintf("[%s/%s] [action:%s]", d.networkChaos.Namespace, d.networkChaos.Name, d.networkChaos.Spec.Action)
}
