// Copyright 2021 Chaos Mesh Authors.
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

package controllers

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
)

type ChaosNodeReconciler struct {
	kubeClient    client.Client
	eventRecorder recorder.ChaosRecorder
	logger        logr.Logger
}

func NewChaosNodeReconciler(kubeClient client.Client, eventRecorder recorder.ChaosRecorder, logger logr.Logger) *ChaosNodeReconciler {
	return &ChaosNodeReconciler{kubeClient: kubeClient, eventRecorder: eventRecorder, logger: logger}
}

func (it *ChaosNodeReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {

	startTime := time.Now()
	defer func() {
		it.logger.V(4).Info("Finished syncing for chaos node",
			"node", request.NamespacedName,
			"duration", time.Since(startTime),
		)
	}()

	ctx := context.TODO()
	node := v1alpha1.WorkflowNode{}

	err := it.kubeClient.Get(ctx, request.NamespacedName, &node)
	if err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	if !v1alpha1.IsChaosTemplateType(node.Spec.Type) {
		return reconcile.Result{}, nil
	}

	it.logger.V(4).Info("resolve chaos node", "node", request)

	if node.Spec.Type == v1alpha1.TypeSchedule {
		err := it.syncSchedule(ctx, node)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else {
		err = it.syncChaosResources(ctx, node)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	updateError := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		nodeNeedUpdate := v1alpha1.WorkflowNode{}
		err := it.kubeClient.Get(ctx, request.NamespacedName, &nodeNeedUpdate)
		if err != nil {
			return client.IgnoreNotFound(err)
		}

		if node.Spec.Type == v1alpha1.TypeSchedule {
			// sync status with schedule
			scheduleList, err := it.fetchChildrenSchedule(ctx, nodeNeedUpdate)
			if err != nil {
				return client.IgnoreNotFound(err)
			}
			if len(scheduleList) > 1 {
				it.logger.Info("the number of schedule custom resource affected by chaos node is more than 1",
					"chaos node", fmt.Sprintf("%s/%s", node.Namespace, node.Name),
					"schedule custom resources", scheduleList,
				)
			}
			if len(scheduleList) > 0 {
				scheduleObject := scheduleList[0]
				group := scheduleObject.GetObjectKind().GroupVersionKind().Group
				chaosRef := corev1.TypedLocalObjectReference{
					APIGroup: &group,
					Kind:     scheduleObject.GetObjectKind().GroupVersionKind().Kind,
					Name:     scheduleObject.GetName(),
				}
				nodeNeedUpdate.Status.ChaosResource = &chaosRef
				SetCondition(&nodeNeedUpdate.Status, v1alpha1.WorkflowNodeCondition{
					Type:   v1alpha1.ConditionChaosInjected,
					Status: corev1.ConditionTrue,
					Reason: v1alpha1.ChaosCRCreated,
				})
			} else {
				nodeNeedUpdate.Status.ChaosResource = nil
				SetCondition(&nodeNeedUpdate.Status, v1alpha1.WorkflowNodeCondition{
					Type:   v1alpha1.ConditionChaosInjected,
					Status: corev1.ConditionFalse,
					Reason: v1alpha1.ChaosCRNotExists,
				})
			}

			return client.IgnoreNotFound(it.kubeClient.Status().Update(ctx, &nodeNeedUpdate))
		}

		// sync status with chaos CustomResource
		chaosList, err := it.fetchChildrenChaosCustomResource(ctx, nodeNeedUpdate)
		if err != nil {
			return client.IgnoreNotFound(err)
		}
		if len(chaosList) > 1 {
			it.logger.Info("the number of chaos custom resource affected by chaos node is more than 1",
				"chaos node", fmt.Sprintf("%s/%s", node.Namespace, node.Name),
				"chaos custom resources", chaosList,
			)
		}

		if len(chaosList) > 0 {
			chaosObject := chaosList[0]
			group := chaosObject.GetObjectKind().GroupVersionKind().Group
			chaosRef := corev1.TypedLocalObjectReference{
				APIGroup: &group,
				Kind:     chaosObject.GetObjectKind().GroupVersionKind().Kind,
				Name:     chaosObject.GetName(),
			}
			nodeNeedUpdate.Status.ChaosResource = &chaosRef
			SetCondition(&nodeNeedUpdate.Status, v1alpha1.WorkflowNodeCondition{
				Type:   v1alpha1.ConditionChaosInjected,
				Status: corev1.ConditionTrue,
				Reason: v1alpha1.ChaosCRCreated,
			})
		} else {
			nodeNeedUpdate.Status.ChaosResource = nil
			SetCondition(&nodeNeedUpdate.Status, v1alpha1.WorkflowNodeCondition{
				Type:   v1alpha1.ConditionChaosInjected,
				Status: corev1.ConditionFalse,
				Reason: v1alpha1.ChaosCRNotExists,
			})
		}

		return client.IgnoreNotFound(it.kubeClient.Status().Update(ctx, &nodeNeedUpdate))
	})

	return reconcile.Result{}, updateError
}

func (it *ChaosNodeReconciler) syncSchedule(ctx context.Context, node v1alpha1.WorkflowNode) error {
	scheduleList, err := it.fetchChildrenSchedule(ctx, node)
	if err != nil {
		return err
	}
	if WorkflowNodeFinished(node.Status) {
		// make the number of schedule to 0
		for _, item := range scheduleList {
			item := item
			err := it.kubeClient.Delete(ctx, &item)
			if client.IgnoreNotFound(err) != nil {
				it.logger.Error(err, "failed to delete schedule CR for workflow chaos node",
					"namespace", node.Namespace,
					"chaos node", node.Name,
					"schedule CR name", item.GetName(),
				)
				it.eventRecorder.Event(&node, recorder.ChaosCustomResourceDeleteFailed{
					Name: item.GetName(),
					Kind: item.GetObjectKind().GroupVersionKind().Kind,
				})
			} else {
				it.eventRecorder.Event(&node, recorder.ChaosCustomResourceDeleted{
					Name: item.GetName(),
					Kind: item.GetObjectKind().GroupVersionKind().Kind,
				})
			}
		}
		return nil
	}
	if len(scheduleList) == 0 {
		return it.createSchedule(ctx, node)
	} else if len(scheduleList) > 1 {
		// need cleanup

		var scheduleCrToRemove []string
		for _, item := range scheduleList[1:] {
			scheduleCrToRemove = append(scheduleCrToRemove, item.GetName())
		}

		it.logger.Info("removing duplicated schedule custom resource",
			"chaos node", fmt.Sprintf("%s/%s", node.Namespace, node.Name),
			"schedule cr to remove", scheduleCrToRemove,
		)

		for _, item := range scheduleList[1:] {
			// best efforts deletion
			item := item
			err := it.kubeClient.Delete(ctx, &item)
			if client.IgnoreNotFound(err) != nil {
				it.logger.Error(err, "failed to delete schedule CR for workflow chaos node",
					"namespace", node.Namespace,
					"chaos node", node.Name,
					"schedule CR name", item.GetName(),
				)
			}
		}
	} else {
		it.logger.V(4).Info("do not need spawn or remove schedule CR")
	}
	return nil

}

func (it *ChaosNodeReconciler) syncChaosResources(ctx context.Context, node v1alpha1.WorkflowNode) error {

	chaosList, err := it.fetchChildrenChaosCustomResource(ctx, node)
	if err != nil {
		return err
	}

	if WorkflowNodeFinished(node.Status) {
		// make the number of chaos resource to 0
		for _, item := range chaosList {
			// best efforts deletion
			item := item
			// TODO: it should not be delete directly with the new implementation of *Chaos controller in branch nirvana
			err := it.kubeClient.Delete(ctx, item)
			if client.IgnoreNotFound(err) != nil {
				it.logger.Error(err, "failed to delete chaos CR for workflow chaos node",
					"namespace", node.Namespace,
					"chaos node", node.Name,
					"chaos CR name", item.GetName(),
				)
				it.eventRecorder.Event(&node, recorder.ChaosCustomResourceDeleteFailed{
					Name: item.GetName(),
					Kind: item.GetObjectKind().GroupVersionKind().Kind,
				})
			} else {
				it.eventRecorder.Event(&node, recorder.ChaosCustomResourceDeleted{
					Name: item.GetName(),
					Kind: item.GetObjectKind().GroupVersionKind().Kind,
				})
			}
		}
		return nil
	}
	// make the number of chaos resource to 1
	if len(chaosList) == 0 {
		return it.createChaos(ctx, node)
	} else if len(chaosList) > 1 {

		var chaosCrToRemove []string
		for _, item := range chaosList[1:] {
			chaosCrToRemove = append(chaosCrToRemove, item.GetName())
		}

		it.logger.Info("removing duplicated chaos custom resource",
			"chaos node", fmt.Sprintf("%s/%s", node.Namespace, node.Name),
			"chaos cr to remove", chaosCrToRemove,
		)

		for _, item := range chaosList[1:] {
			// best efforts deletion
			item := item
			err := it.kubeClient.Delete(ctx, item)
			if client.IgnoreNotFound(err) != nil {
				it.logger.Error(err, "failed to delete chaos CR for workflow chaos node",
					"namespace", node.Namespace,
					"chaos node", node.Name,
					"chaos CR name", item.GetName(),
				)
			}
		}
	} else {
		it.logger.V(4).Info("do not need spawn or remove chaos CR")
	}

	// TODO: also respawn the chaos resource if Spec changed in workflow

	return nil
}

// inject Chaos will create one instance of chaos CR
func (it *ChaosNodeReconciler) createChaos(ctx context.Context, node v1alpha1.WorkflowNode) error {

	chaosObject, meta, err := node.Spec.EmbedChaos.SpawnNewObject(node.Spec.Type)
	if err != nil {
		return err
	}

	meta.SetGenerateName(fmt.Sprintf("%s-", node.Name))
	meta.SetNamespace(node.Namespace)
	meta.SetOwnerReferences(append(meta.GetOwnerReferences(), metav1.OwnerReference{
		APIVersion:         node.APIVersion,
		Kind:               node.Kind,
		Name:               node.Name,
		UID:                node.UID,
		Controller:         &isController,
		BlockOwnerDeletion: &blockOwnerDeletion,
	}))
	meta.SetLabels(map[string]string{
		v1alpha1.LabelControlledBy: node.Name,
		v1alpha1.LabelWorkflow:     node.Spec.WorkflowName,
	})

	err = it.kubeClient.Create(ctx, chaosObject)
	if err != nil {
		it.eventRecorder.Event(&node, recorder.ChaosCustomResourceCreateFailed{})
		it.logger.Error(err, "failed to create chaos")
		return nil
	}
	it.logger.Info("chaos object created", "namespace", meta.GetNamespace(), "name", meta.GetName())
	it.eventRecorder.Event(&node, recorder.ChaosCustomResourceCreated{
		Name: meta.GetName(),
		Kind: chaosObject.GetObjectKind().GroupVersionKind().Kind,
	})
	return nil
}

func (it *ChaosNodeReconciler) fetchChildrenChaosCustomResource(ctx context.Context, node v1alpha1.WorkflowNode) ([]v1alpha1.GenericChaos, error) {
	genericChaosList, err := node.Spec.EmbedChaos.SpawnNewList(node.Spec.Type)
	if err != nil {
		return nil, err
	}
	controlledByThisNode, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: map[string]string{
			v1alpha1.LabelControlledBy: node.Name,
		},
	})
	if err != nil {
		it.logger.Error(err, "failed to build label selector with filtering children workflow node controlled by current node",
			"current node", fmt.Sprintf("%s/%s", node.Namespace, node.Name))
		return nil, err
	}

	err = it.kubeClient.List(ctx, genericChaosList, &client.ListOptions{
		LabelSelector: controlledByThisNode,
	})
	if err != nil {
		return nil, err
	}

	var sorted SortGenericChaosByCreationTimestamp = genericChaosList.GetItems()
	sort.Sort(sorted)
	return sorted, err
}

func (it ChaosNodeReconciler) createSchedule(ctx context.Context, node v1alpha1.WorkflowNode) error {
	if node.Spec.Schedule == nil {
		return fmt.Errorf("invalid workfow node, the spec of schedule is nil")
	}
	scheduleToCreate := v1alpha1.Schedule{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    node.Namespace,
			GenerateName: fmt.Sprintf("%s-", node.Name),
			Labels: map[string]string{
				v1alpha1.LabelControlledBy: node.Name,
				v1alpha1.LabelWorkflow:     node.Spec.WorkflowName,
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         node.APIVersion,
					Kind:               node.Kind,
					Name:               node.Name,
					UID:                node.UID,
					Controller:         &isController,
					BlockOwnerDeletion: &blockOwnerDeletion,
				},
			},
		},
		Spec: *node.Spec.Schedule,
	}
	err := it.kubeClient.Create(ctx, &scheduleToCreate)
	if err != nil {
		it.eventRecorder.Event(&node, recorder.ChaosCustomResourceCreateFailed{})
		it.logger.Error(err, "failed to create schedule CR")
		return nil
	}
	it.logger.Info("schedule CR created", "namespace", scheduleToCreate.GetNamespace(), "name", scheduleToCreate.GetName())
	it.eventRecorder.Event(&node, recorder.ChaosCustomResourceCreated{
		Name: scheduleToCreate.GetName(),
		Kind: scheduleToCreate.GetObjectKind().GroupVersionKind().Kind,
	})
	return nil

}

func (it *ChaosNodeReconciler) fetchChildrenSchedule(ctx context.Context, node v1alpha1.WorkflowNode) ([]v1alpha1.Schedule, error) {
	var scheduleList v1alpha1.ScheduleList
	controlledByThisNode, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: map[string]string{
			v1alpha1.LabelControlledBy: node.Name,
		},
	})
	if err != nil {
		it.logger.Error(err, "failed to build label selector with filtering children workflow node controlled by current node",
			"current node", fmt.Sprintf("%s/%s", node.Namespace, node.Name))
		return nil, err
	}
	err = it.kubeClient.List(ctx, &scheduleList, &client.ListOptions{
		LabelSelector: controlledByThisNode,
	})
	if err != nil {
		return nil, err
	}
	var sorted SortScheduleByCreationTimestamp = scheduleList.Items
	sort.Sort(sorted)
	return sorted, err
}

type SortGenericChaosByCreationTimestamp []v1alpha1.GenericChaos

func (it SortGenericChaosByCreationTimestamp) Len() int {
	return len(it)
}

func (it SortGenericChaosByCreationTimestamp) Less(i, j int) bool {
	return it[j].GetCreationTimestamp().After(it[i].GetCreationTimestamp().Time)
}

func (it SortGenericChaosByCreationTimestamp) Swap(i, j int) {
	it[i], it[j] = it[j], it[i]
}

type SortScheduleByCreationTimestamp []v1alpha1.Schedule

func (it SortScheduleByCreationTimestamp) Len() int {
	return len(it)
}

func (it SortScheduleByCreationTimestamp) Less(i, j int) bool {
	return it[j].GetCreationTimestamp().After(it[i].GetCreationTimestamp().Time)
}

func (it SortScheduleByCreationTimestamp) Swap(i, j int) {
	it[i], it[j] = it[j], it[i]
}
