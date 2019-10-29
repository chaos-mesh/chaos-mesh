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

// TODO: some of these codes are copied directly from podchaos. Refractor is needed for reusing code

package networkchaos

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	"github.com/juju/errors"
	"github.com/pingcap/chaos-operator/pkg/apis/pingcap.com/v1alpha1"
	"github.com/pingcap/chaos-operator/pkg/client/clientset/versioned"
	informers "github.com/pingcap/chaos-operator/pkg/client/informers/externalversions"
	listers "github.com/pingcap/chaos-operator/pkg/client/listers/pingcap.com/v1alpha1"
	"github.com/pingcap/chaos-operator/pkg/manager"

	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	eventv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

const (
	// SuccessSynced is used as part of the Event 'reason' when a NetworkChaos is synced
	SuccessSynced = "Synced"

	// MessageResourceSynced is the message used for an Event fired when a NetworkChaos
	// is synced successfully
	MessageResourceSynced = "PodChaos synced successfully"
)

// ControlInterface implements the control logic for controlling PodChaos
// It is implemented as an interface to allow for extensions that provide different semantics.
type ControlInterface interface {
	UpdateNetworkChaos(networkChaos *v1alpha1.NetworkChaos) error
	DeleteNetworkChaos(key string) error
}

// Controller is the controller implementation for network chaos resources.
type Controller struct {
	// kubernetes client interface
	kubeCli kubernetes.Interface
	// operator client interface
	cli versioned.Interface
	// control returns an interface capable of syncing a podchaos object.
	control ControlInterface
	// PodLister is able to list/get pod object from a shred informers's store
	podLister corelisters.PodLister
	// podsSynced returns true if the pod object shared informer has synced at least once.
	podsSynced cache.InformerSynced
	// ncLister is able to list/get networkchaos object from a shared informer's store.
	ncLister listers.NetworkChaosLister
	// ncsSynced returns true if the networkchaos object shared informer has synced at least once.
	ncsSynced cache.InformerSynced
	// queue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	queue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

// NewController returns a new network chaos controller.
func NewController(
	kubeCli kubernetes.Interface,
	cli versioned.Interface,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	informerFactory informers.SharedInformerFactory,
	managerBase manager.ManagerBaseInterface,
) *Controller {
	// Create event broadcaster.
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&eventv1.EventSinkImpl{
		Interface: eventv1.New(kubeCli.CoreV1().RESTClient()).Events("")})
	recorder := eventBroadcaster.NewRecorder(v1alpha1.Scheme, corev1.EventSource{Component: "networkChaos"})

	ncInformer := informerFactory.Pingcap().V1alpha1().NetworkChaoses()
	podInformer := kubeInformerFactory.Core().V1().Pods()

	controller := &Controller{
		kubeCli:    kubeCli,
		cli:        cli,
		podLister:  podInformer.Lister(),
		podsSynced: podInformer.Informer().HasSynced,
		ncLister:   ncInformer.Lister(),
		ncsSynced:  ncInformer.Informer().HasSynced,
		queue:      workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "NetworkChaos"),
		recorder:   recorder,
	}

	controller.control = NewNetworkChaosControl(
		kubeCli,
		cli,
		managerBase,
		controller.podLister,
		controller.ncLister,
	)

	glog.Info("Setting up network chaos event handlers")

	ncInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueNetworkChaos,
		UpdateFunc: func(old, cur interface{}) {
			controller.enqueueNetworkChaos(cur)
		},
		DeleteFunc: controller.enqueueNetworkChaos,
	})

	return controller
}

// Run runs the network chaos controller.
func (c *Controller) Run(stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	glog.Info("Starting network chaos controller")

	glog.Info("Waiting informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.ncsSynced, c.podsSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sysn")
	}

	go wait.Until(c.worker, time.Second, stopCh)

	<-stopCh
	glog.Info("Shutting network chaos controller")

	return nil
}

// worker runs a worker goroutine that invokes processNextWorkItem until the the controller's queue is closed
func (c *Controller) worker() {
	for c.processNextWorkItem() {
		// revive:disable:empty-block
	}
}

// processNextWorkItem dequeues items, processes them, and marks them done. It enforces that the syncHandler is never
// invoked concurrently with the same key.
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.queue.Get()
	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.queue.Done(obj)
		var (
			key string
			ok  bool
		)
		if key, ok = obj.(string); !ok {
			// As the item in the queue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.queue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}

		if err := c.syncHandler(key); err != nil {
			// Put the item back on the queue to handle any transient errors.
			c.queue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}

		c.queue.Forget(obj)
		glog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

// syncHandler sync the given NetworkChaos.
func (c *Controller) syncHandler(key string) error {
	startTime := time.Now()
	defer func() {
		glog.V(4).Infof("Finished syncing NetworkChaos %q (%v)", key, time.Since(startTime))
	}()

	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	nc, err := c.ncLister.NetworkChaoses(ns).Get(name)
	// The NetworkChaos may no longer exist, in which case we think the networkChaos has been deleted.
	if errors.IsNotFound(err) {
		utilruntime.HandleError(fmt.Errorf("networkChaos '%s' in work queue no longer exists", key))
		return c.control.DeleteNetworkChaos(key)
	}

	if err != nil {
		return err
	}

	if err := c.syncNetworkChaos(nc); err != nil {
		return err
	}

	c.recorder.Event(nc, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (c *Controller) syncNetworkChaos(nc *v1alpha1.NetworkChaos) error {
	return c.control.UpdateNetworkChaos(nc)
}

// enqueueNetworkChaos takes a NetworkChaos resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than NetworkChaos.
func (c *Controller) enqueueNetworkChaos(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)

	if err != nil {
		utilruntime.HandleError(fmt.Errorf("cound't get key for object %+v: %v", obj, err))
		return
	}

	c.queue.Add(key)
}
