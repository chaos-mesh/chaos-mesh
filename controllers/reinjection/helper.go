// Copyright 2023 Chaos Mesh Authors.
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

package reinjection

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	chaosmeshapi "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

type chaosKeyInfo struct {
	Kind      string
	Name      string
	Namespace string
}

func buildPodPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc:  func(e event.CreateEvent) bool { return false },
		GenericFunc: func(e event.GenericEvent) bool { return false },
		DeleteFunc:  func(e event.DeleteEvent) bool { return false },
		UpdateFunc: func(e event.UpdateEvent) bool {
			// find restarted container name
			oldPod := e.ObjectOld.(*corev1.Pod)
			podKey := buildPodKey(oldPod)
			chaosToContainerMapValue, ok := podToChaosInfoMap.Load(podKey)
			if !ok {
				return false
			}
			newPod := e.ObjectNew.(*corev1.Pod)
			restartedContainerNames := make([]string, 0)
			oldContainers := make(map[string]corev1.ContainerStatus)
			newContainers := make(map[string]corev1.ContainerStatus)
			for _, container := range oldPod.Status.ContainerStatuses {
				oldContainers[container.Name] = container
			}
			for _, container := range newPod.Status.ContainerStatuses {
				newContainers[container.Name] = container
			}
			for _, container := range newContainers {
				if oldContainer, ok := oldContainers[container.Name]; ok {
					if oldContainer.RestartCount != container.RestartCount {
						restartedContainerNames = append(restartedContainerNames, container.Name)
					}
				}
			}
			if len(restartedContainerNames) == 0 {
				return false
			}

			// if we get here, it means that the pod has restarted containers
			chaosToContainerMap := chaosToContainerMapValue.(map[string]string)
			injectedContainers := make(map[string]bool)
			for _, containerName := range chaosToContainerMap {
				// this means that this kind of chaos is pod scoped,
				// so no need to find the container name, just trigger the re-injection
				if containerName == "" {
					return true
				}
				injectedContainers[containerName] = true
			}
			for _, name := range restartedContainerNames {
				if _, ok = injectedContainers[name]; ok {
					return true
				}
			}
			return false
		},
	}
}

func buildChaosPredicate(chaosKind string) predicate.Predicate {
	return predicate.Funcs{
		CreateFunc:  func(e event.CreateEvent) bool { return false },
		GenericFunc: func(e event.GenericEvent) bool { return false },
		UpdateFunc: func(e event.UpdateEvent) bool {
			newObj := e.ObjectNew.(chaosmeshapi.StatefulObject)
			newStatus := newObj.GetStatus()
			if newStatus.Conditions == nil {
				return false
			}
			for _, condition := range newStatus.Conditions {
				// if the chaos is time exceeded, we should remove the chaos from the map
				if newStatus.Experiment.DesiredPhase == chaosmeshapi.StoppedPhase {
					removeChaosFromPodToChaosInfoMap(chaosKind, newObj)
					return false
				}

				if (condition.Type == chaosmeshapi.ConditionSelected && condition.Status == corev1.ConditionFalse) ||
					(condition.Type == chaosmeshapi.ConditionAllInjected && condition.Status == corev1.ConditionFalse) {
					return false
				}
			}

			addChaosToPodToChaosInfoMap(chaosKind, newObj)
			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			chaos := e.Object.(chaosmeshapi.StatefulObject)
			removeChaosFromPodToChaosInfoMap(chaosKind, chaos)
			return false
		},
	}
}

func addChaosToPodToChaosInfoMap(chaosKind string, chaos chaosmeshapi.StatefulObject) {
	chaosKey := buildChaosKey(chaosKind, chaos)
	status := chaos.GetStatus()
	for _, record := range status.Experiment.Records {
		podKey, containerName, err := parseRecordId(record.Id)
		if err != nil {
			continue
		}
		chaosToContainerMapValue, ok := podToChaosInfoMap.Load(podKey)
		var chaosToContainerMap map[string]string
		if !ok {
			chaosToContainerMap = make(map[string]string)
			chaosToContainerMap[chaosKey] = containerName
		} else {
			chaosToContainerMap = chaosToContainerMapValue.(map[string]string)
			chaosToContainerMap[chaosKey] = containerName
		}
		podToChaosInfoMap.Store(podKey, chaosToContainerMap)
	}
}

func removeChaosFromPodToChaosInfoMap(chaosKind string, chaos chaosmeshapi.StatefulObject) {
	status := chaos.GetStatus()
	chaosKey := buildChaosKey(chaosKind, chaos)
	for _, record := range status.Experiment.Records {
		podKey, _, err := parseRecordId(record.Id)
		if err != nil {
			continue
		}
		chaosToContainerMapValue, ok := podToChaosInfoMap.Load(podKey)
		if !ok {
			continue
		}
		chaosToContainerMap := chaosToContainerMapValue.(map[string]string)
		delete(chaosToContainerMap, chaosKey)
		if len(chaosToContainerMap) == 0 {
			podToChaosInfoMap.Delete(podKey)
		} else {
			podToChaosInfoMap.Store(podKey, chaosToContainerMap)
		}
	}
}

func buildPodKey(pod *corev1.Pod) string {
	return fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
}

func buildChaosKey(chaosKind string, chaos chaosmeshapi.StatefulObject) string {
	return fmt.Sprintf("%s/%s/%s", chaosKind, chaos.GetNamespace(), chaos.GetName())
}

func parseChaosKey(key string) (*chaosKeyInfo, error) {
	parts := strings.Split(key, "/")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid chaos key %s", key)
	}
	return &chaosKeyInfo{
		Kind:      parts[0],
		Namespace: parts[1],
		Name:      parts[2],
	}, nil
}

func parseRecordId(recordId string) (podKey, containerName string, err error) {
	parts := strings.Split(recordId, "/")
	if len(parts) == 3 {
		podKey = fmt.Sprintf("%s/%s", parts[0], parts[1])
		containerName = parts[2]
		return podKey, containerName, nil
	}
	if len(parts) == 2 {
		return recordId, "", nil
	}
	return "", "", fmt.Errorf("invalid record id %s", recordId)
}

func syncChaos(client client.Client, controllerConfig *ControllerConfig) error {
	for _, chaosKindString := range controllerConfig.ChaosKinds {
		chaosKind, exists := chaosmeshapi.AllKinds()[chaosKindString]
		if !exists {
			continue
		}
		objects := chaosKind.SpawnList()
		if err := client.List(context.Background(), objects); err != nil {
			return err
		}

		for _, chaos := range objects.GetItems() {
			statefulObj := chaos.(chaosmeshapi.StatefulObject)
			status := statefulObj.GetStatus()
			if status.Conditions == nil {
				continue
			}
			for _, condition := range status.Conditions {
				// if the chaos is time exceeded, we should NOT put it into podToChaosInfoMap
				if status.Experiment.DesiredPhase == chaosmeshapi.StoppedPhase {
					continue
				}

				// if the chaos is not selected or not injected, we should NOT put it into podToChaosInfoMap
				if (condition.Type == chaosmeshapi.ConditionSelected && condition.Status == corev1.ConditionFalse) ||
					(condition.Type == chaosmeshapi.ConditionAllInjected && condition.Status == corev1.ConditionFalse) {
					continue
				}
			}
			addChaosToPodToChaosInfoMap(chaosKindString, statefulObj)
		}
	}
	return nil
}
