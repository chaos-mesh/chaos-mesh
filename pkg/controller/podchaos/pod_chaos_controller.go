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

package podchaos

import (
	"fmt"
	"time"

	"github.com/cwen0/chaos-operator/pkg/apis/pingcap.com/v1alpha1"
	"github.com/cwen0/chaos-operator/pkg/client/clientset/versioned"
	informers "github.com/cwen0/chaos-operator/pkg/client/informers/externalversions"
	listers "github.com/cwen0/chaos-operator/pkg/client/listers/pingcap.com/v1alpha1"
	"github.com/golang/glog"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	eventv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

const (
	// SuccessSynced is used as part of the Event 'reason' when a PodChaos is synced
	SuccessSynced = "Synced"

	// MessageResourceSynced is the message used for an Event fired when a PodChaos
	// is synced successfully
	MessageResourceSynced = "PodChaos synced successfully"
)

// ControlInterface implements the control logic for updating PodChaos
// It is implemented as an interface to allow for extensions that provide different semantics.
type ControlInterface interface {
	UpdatePodChaos(podChaos *v1alpha1.PodChaos) error
}

// Controller is the controller implementation for pod chaos resources.
type Controller struct {
	// kubernetes client interface
	kubeCli kubernetes.Interface
	// operator client interface
	cli versioned.Interface
	// control returns an interface capable of syncing a podchaos object.
	control ControlInterface
	// pcLister is able to list/get podchaos object from a shared informer's store.
	pcLister listers.PodChaosLister
	// PcListerSynced returns true if the podchaos object shared informer has synced at least once.
	pcListerSynced cache.InformerSynced
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

// NewController returns a new pod chaos controller.
func NewController(
	kubeCli kubernetes.Interface,
	cli versioned.Interface,
	_ kubeinformers.SharedInformerFactory,
	informerFactory informers.SharedInformerFactory,
) *Controller {
	// Create event broadcaster.
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&eventv1.EventSinkImpl{
		Interface: eventv1.New(kubeCli.CoreV1().RESTClient()).Events("")})
	recorder := eventBroadcaster.NewRecorder(v1alpha1.Scheme, corev1.EventSource{Component: "podChaos"})

	pcInformer := informerFactory.Pingcap().V1alpha1().PodChaoses()
	controller := &Controller{
		kubeCli:        kubeCli,
		cli:            cli,
		control:        NewPodChaosControl(),
		pcLister:       pcInformer.Lister(),
		pcListerSynced: pcInformer.Informer().HasSynced,
		queue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "PodChaos"),
		recorder:       recorder,
	}

	glog.Info("Setting up pod chaos event handlers")

	pcInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueuePodChaos,
		UpdateFunc: func(old, cur interface{}) {
			controller.enqueuePodChaos(cur)
		},
		DeleteFunc: controller.enqueuePodChaos,
	})

	return controller
}

// Run runs the podchaos controller.
func (c *Controller) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	glog.Info("Starting pod chaos controller")
	defer glog.Info("Shutting pod chaos controller")

	for i := 0; i < workers; i++ {
		go wait.Until(c.worker, time.Second, stopCh)
	}

	<-stopCh
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

// syncHandler sync the given PodChaos.
func (c *Controller) syncHandler(key string) error {
	startTime := time.Now()
	defer func() {
		glog.V(4).Infof("Finished syncing PodChaos %q (%v)", key, time.Since(startTime))
	}()

	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	pc, err := c.pcLister.PodChaoses(ns).Get(name)
	if errors.IsNotFound(err) {
		glog.Infof("PodChaos has been deleted %v", key)
		return nil
	}

	if err != nil {
		return err
	}

	if err := c.syncPodChaos(pc); err != nil {
		return err
	}

	c.recorder.Event(pc, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (c *Controller) syncPodChaos(pc *v1alpha1.PodChaos) error {
	return c.control.UpdatePodChaos(pc)
}

// enqueuePodChaos takes a PodChaos resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than PodChaos.
func (c *Controller) enqueuePodChaos(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("cound't get key for object %+v: %v", obj, err))
		return
	}

	c.queue.Add(key)
}
