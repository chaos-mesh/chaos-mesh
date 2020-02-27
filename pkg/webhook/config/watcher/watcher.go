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

package watcher

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/pingcap/chaos-mesh/pkg/webhook/config"

	ctrl "sigs.k8s.io/controller-runtime"

	v1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	k8sv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

var log = ctrl.Log.WithName("inject-webhook")
var restInClusterConfig = rest.InClusterConfig
var kubernetesNewForConfig = kubernetes.NewForConfig

const (
	serviceAccountNamespaceFilePath = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
)

// ErrWatchChannelClosed should restart watcher
var ErrWatchChannelClosed = errors.New("watcher channel has closed")

// K8sConfigMapWatcher is a struct that connects to the API and collects, parses, and emits sidecar configurations
type K8sConfigMapWatcher struct {
	Config
	client k8sv1.CoreV1Interface
}

// New creates a new K8sConfigMapWatcher
func New(cfg Config) (*K8sConfigMapWatcher, error) {
	c := K8sConfigMapWatcher{Config: cfg}
	if c.Namespace == "" {
		// ENHANCEMENT: support downward API/env vars instead? https://github.com/kubernetes/kubernetes/blob/release-1.0/docs/user-guide/downward-api.md
		// load from file on disk for serviceaccount: /var/run/secrets/kubernetes.io/serviceaccount/namespace
		ns, err := ioutil.ReadFile(serviceAccountNamespaceFilePath)
		if err != nil {
			return nil, fmt.Errorf("%s: maybe you should specify --configmap-namespace if you are running outside of kubernetes", err.Error())
		}
		if string(ns) != "" {
			c.Namespace = string(ns)
			log.Info("Inferred ConfigMap",
				"namespace", c.Namespace, "filepath", serviceAccountNamespaceFilePath)
		}
	}

	log.Info("Creating Kubernetes client from in-cluster discovery")
	k8sConfig, err := restInClusterConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetesNewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	c.client = clientset.CoreV1()
	err = validate(&c)
	if err != nil {
		return nil, fmt.Errorf("validation failed for K8sConfigMapWatcher: %s", err.Error())
	}
	log.Info("Created ConfigMap watcher",
		"apiserver", k8sConfig.Host, "namespaces", c.Namespace, "watchlabels", c.ConfigMapLabels)
	return &c, nil
}

func validate(c *K8sConfigMapWatcher) error {
	if c == nil {
		return fmt.Errorf("configmap watcher was nil")
	}
	if c.Namespace == "" {
		return fmt.Errorf("namespace is empty")
	}
	if c.ConfigMapLabels == nil {
		return fmt.Errorf("configmap labels was an uninitialized map")
	}
	if c.client == nil {
		return fmt.Errorf("k8s client was not setup properly")
	}
	return nil
}

// Watch watches for events impacting watched ConfigMaps and emits their events across a channel
func (c *K8sConfigMapWatcher) Watch(notifyMe chan<- interface{}, stopCh <-chan struct{}) error {
	log.Info("Watching for ConfigMaps for changes",
		"namespace", c.Namespace, "labels", c.ConfigMapLabels)
	watcher, err := c.client.ConfigMaps(c.Namespace).Watch(metav1.ListOptions{
		LabelSelector: mapStringStringToLabelSelector(c.ConfigMapLabels),
	})
	if err != nil {
		return fmt.Errorf("unable to create watcher (possible serviceaccount RBAC/ACL failure?): %s", err.Error())
	}
	defer watcher.Stop()
	for {
		select {
		case e, ok := <-watcher.ResultChan():
			// channel may closed caused by HTTP timeout, should restart watcher
			// detail at https://github.com/kubernetes/client-go/issues/334
			if !ok {
				log.Error(nil, "channel has closed, should restart watcher")
				return ErrWatchChannelClosed
			}
			if e.Type == watch.Error {
				return apierrs.FromObject(e.Object)
			}
			log.V(3).Info("type", e.Type, "kind", e.Object.GetObjectKind())
			switch e.Type {
			case watch.Added:
				fallthrough
			case watch.Modified:
				fallthrough
			case watch.Deleted:
				// signal reconciliation of all InjectionConfigs
				log.V(3).Info("Signalling event received from watch channel",
					"type", e.Type, "kind", e.Object.GetObjectKind())
				notifyMe <- struct{}{}
			default:
				log.Error(nil, "got unsupported event! skipping", "type", e.Type, "kind", e.Object.GetObjectKind())
			}
			// events! yay!
		case <-stopCh:
			log.V(2).Info("Stopping configmap watcher, context indicated we are done")
			// clean up, we cancelled the context, so stop the watch
			return nil
		}
	}
}

func mapStringStringToLabelSelector(m map[string]string) string {
	// https://github.com/kubernetes/apimachinery/issues/47
	return labels.Set(m).String()
}

// Get fetches all matching ConfigMaps
func (c *K8sConfigMapWatcher) Get() (cfgs []*config.InjectionConfig, err error) {
	log.Info("Fetching ConfigMaps...")
	clist, err := c.client.ConfigMaps(c.Namespace).List(metav1.ListOptions{
		LabelSelector: mapStringStringToLabelSelector(c.ConfigMapLabels),
	})
	if err != nil {
		return cfgs, err
	}

	if clist == nil {
		return cfgs, nil
	}

	log.Info("Fetched ConfigMaps", "configmap count", len(clist.Items))
	for _, cm := range clist.Items {
		injectionConfigsForCM, err := InjectionConfigsFromConfigMap(cm)
		if err != nil {
			return cfgs, fmt.Errorf("error getting ConfigMaps from API: %s", err.Error())
		}
		log.V(1).Info("Found InjectionConfigs",
			"count", len(injectionConfigsForCM), "name", cm.ObjectMeta.Name)
		cfgs = append(cfgs, injectionConfigsForCM...)
	}
	return cfgs, nil
}

// InjectionConfigsFromConfigMap parse items in a configmap into a list of InjectionConfigs
func InjectionConfigsFromConfigMap(cm v1.ConfigMap) ([]*config.InjectionConfig, error) {
	ics := []*config.InjectionConfig{}
	for name, payload := range cm.Data {
		log.Info("Parsing InjectionConfig",
			"namespace", cm.ObjectMeta.Namespace, "meta name", cm.ObjectMeta.Name, "config name", name)
		ic, err := config.LoadInjectionConfig(strings.NewReader(payload))
		if err != nil {
			return nil, fmt.Errorf("error parsing ConfigMap %s item %s into injection config: %s", cm.ObjectMeta.Name, name, err.Error())
		}
		log.Info("Loaded InjectionConfig from ConfigMap",
			"config name", ic.Name, "meta name", cm.ObjectMeta.Name, "config name", name)
		ics = append(ics, ic)
	}
	return ics, nil
}
