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

package collector

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-logr/logr"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	"github.com/pingcap/chaos-mesh/pkg/apiinterface"
	"github.com/pingcap/chaos-mesh/pkg/utils"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ChaosCollector struct {
	client.Client
	Log            logr.Logger
	apiType        runtime.Object
	databaseClient *DatabaseClient
}

func (r *ChaosCollector) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	if r.apiType == nil {
		r.Log.Error(nil, "apiType has not been initialized")
		return ctrl.Result{}, nil
	}
	ctx := context.Background()

	obj, ok := r.apiType.DeepCopyObject().(apiinterface.StatefulObject)
	if !ok {
		r.Log.Error(nil, "it's not a stateful object")
		return ctrl.Result{}, nil
	}
	if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
		r.Log.Error(err, "unable to get chaos")
		return ctrl.Result{}, nil
	}

	status := obj.GetStatus()

	affectedNamespace := make(map[string]bool)
	for _, pod := range status.Experiment.Pods {
		affectedNamespace[pod.Namespace] = true
	}

	for namespace := range affectedNamespace {
		err := r.EnsureTidbNamespaceHasGrafana(ctx, namespace)
		if err != nil {
			r.Log.Error(err, "check grafana for tidb cluster failed")
		}
	}

	if status.Experiment.Phase == v1alpha1.ExperimentPhaseRunning {
		event := Event{
			Name:              req.Name,
			Namespace:         req.Namespace,
			Type:              reflect.TypeOf(obj).Elem().Name(),
			AffectedNamespace: affectedNamespace,
			StartTime:         &status.Experiment.StartTime.Time,
			EndTime:           nil,
		}
		r.Log.Info("Event started, save to database", "event", event)

		err := r.databaseClient.WriteEvent(event)
		if err != nil {
			r.Log.Error(err, "write event to database error")
			return ctrl.Result{}, nil
		}
	} else if status.Experiment.Phase == v1alpha1.ExperimentPhaseFinished || status.Experiment.Phase == v1alpha1.ExperimentPhasePaused {
		event := Event{
			Name:              req.Name,
			Namespace:         req.Namespace,
			Type:              reflect.TypeOf(obj).Elem().Name(),
			AffectedNamespace: affectedNamespace,
			StartTime:         &status.Experiment.StartTime.Time,
			EndTime:           &status.Experiment.EndTime.Time,
		}
		r.Log.Info("Event finished, save to database", "event", event)

		err := r.databaseClient.UpdateEvent(event)
		if err != nil {
			r.Log.Error(err, "write event to database error")
			return ctrl.Result{}, nil
		}
	}
	return ctrl.Result{}, nil
}

func (r *ChaosCollector) Setup(mgr ctrl.Manager, apiType runtime.Object) error {
	r.apiType = apiType

	databaseClient, err := NewDatabaseClient(utils.DataSource)
	if err != nil {
		r.Log.Error(err, "create database client failed")
		return nil
	}

	r.databaseClient = databaseClient
	return ctrl.NewControllerManagedBy(mgr).
		For(apiType).
		Complete(r)
}

func (r *ChaosCollector) EnsureTidbNamespaceHasGrafana(ctx context.Context, namespace string) error {
	var svcList corev1.ServiceList

	var listOptions = client.ListOptions{}
	listOptions.Namespace = namespace
	err := r.List(ctx, &svcList, &listOptions)
	if err != nil {
		r.Log.Error(err, "error while getting all services", "namespace", namespace)
	}

	for _, service := range svcList.Items {
		if strings.Contains(service.Name, "prometheus") {
			ok, err := r.IsGrafanaSetUp(ctx, service.Name, service.Namespace)
			if err != nil {
				r.Log.Error(err, "error while getting grafana")
				return err
			}

			if !ok {
				err := r.SetupGrafana(ctx, service.Name, service.Namespace, service.Spec.Ports[0].Port) // This zero index is unsafe hack. TODO: use a better way to get port
				if err != nil {
					r.Log.Error(err, "error while creating grafana")
					return err
				}
				r.Log.Info("create grafana successfully", "name", service.Name, "namespace", service.Namespace, "port", service.Spec.Ports[0].Port) // This zero index is unsafe hack TODO: use a better way to get port
			}
			break
		}
	}

	return nil
}

func (r *ChaosCollector) IsGrafanaSetUp(ctx context.Context, name string, namespace string) (bool, error) {
	var deploymentList v1.DeploymentList

	var listOptions = client.ListOptions{}
	listOptions.Namespace = utils.DashboardNamespace
	err := r.List(ctx, &deploymentList, &listOptions)
	if err != nil {
		r.Log.Error(err, "error while getting all deployments", "namespace", utils.DashboardNamespace)
	}

	result := false
	for _, deployment := range deploymentList.Items {
		if strings.Contains(deployment.Name, namespace) && strings.Contains(deployment.Name, "-chaos-grafana") {
			result = true
		}
	}

	return result, nil
}

func (r *ChaosCollector) SetupGrafana(ctx context.Context, name string, namespace string, port int32) error {
	var deployment v1.Deployment

	deployment.Namespace = utils.DashboardNamespace
	deployment.Name = fmt.Sprintf("%s-chaos-grafana", namespace)

	labels := map[string]string{
		"app.kubernetes.io/name":      deployment.Name,
		"app.kubernetes.io/component": "grafana",
		"prometheus/name":             name,
		"prometheus/namespace":        namespace,
	}

	var chaosDashboard v1.Deployment
	err := r.Get(ctx, types.NamespacedName{
		Namespace: utils.DashboardNamespace,
		Name:      "chaos-dashboard",
	}, &chaosDashboard)
	if err != nil {
		return err
	}
	uid := chaosDashboard.UID

	deployment.Labels = labels
	deployment.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: labels,
	}
	deployment.Spec.Template.Labels = labels
	blockOwnerDeletion := true
	deployment.OwnerReferences = append(deployment.OwnerReferences, metav1.OwnerReference{
		BlockOwnerDeletion: &blockOwnerDeletion,
		Name:               "chaos-dashboard",
		Kind:               "Deployment",
		APIVersion:         "apps/v1",
		UID:                uid,
	})

	var container corev1.Container
	container.Name = "grafana"
	container.Image = "pingcap/chaos-grafana:latest"
	container.ImagePullPolicy = corev1.PullIfNotPresent
	container.Env = []corev1.EnvVar{
		{
			Name:  "CHAOS_NS",
			Value: namespace,
		},
		{
			Name:  "CHAOS_EVENT_DS_URL",
			Value: fmt.Sprintf("chaos-collector-database.%s:3306", utils.DashboardNamespace),
		},
		{
			Name:  "CHAOS_EVENT_DS_DB",
			Value: "chaos_operator",
		},
		{
			Name:  "CHAOS_EVENT_DS_USER",
			Value: "root",
		},
		{
			Name:  "CHAOS_METRIC_DS_URL",
			Value: fmt.Sprintf("http://%s.%s:%d", name, namespace, port),
		},
	}
	deployment.Spec.Template.Spec.Containers = append(deployment.Spec.Template.Spec.Containers, container)

	r.Log.Info("Creating grafana deployments")
	err = r.Create(ctx, &deployment)
	if err != nil {
		return err
	}

	var service corev1.Service
	service.Name = fmt.Sprintf("%s-chaos-grafana", namespace)
	service.Namespace = utils.DashboardNamespace
	service.Labels = labels
	service.OwnerReferences = append(service.OwnerReferences, metav1.OwnerReference{
		BlockOwnerDeletion: &blockOwnerDeletion,
		Name:               "chaos-dashboard",
		Kind:               "Deployment",
		APIVersion:         "apps/v1",
		UID:                uid,
	})
	service.Spec.Selector = labels
	service.Spec.Ports = []corev1.ServicePort{
		{
			Protocol: corev1.ProtocolTCP,
			Port:     3000,
			TargetPort: intstr.IntOrString{
				IntVal: 3000,
			},
		},
	}
	r.Log.Info("Creating grafana service")
	return r.Create(ctx, &service)
}
