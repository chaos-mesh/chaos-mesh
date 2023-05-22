package reinjection

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"strings"
	"time"

	chaosmeshapi "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type chaosKeyInfo struct {
	Kind      string
	Name      string
	Namespace string
}

func buildPodPredicate(ctx context.Context, cli client.Client) predicate.Predicate {
	return predicate.Funcs{
		CreateFunc:  func(e event.CreateEvent) bool { return false },
		GenericFunc: func(e event.GenericEvent) bool { return false },
		DeleteFunc:  func(e event.DeleteEvent) bool { return false },
		UpdateFunc: func(e event.UpdateEvent) bool {
			// find restarted container name
			oldPod := e.ObjectOld.(*corev1.Pod)
			podKey := buildPodKey(oldPod)
			chaosToContainerMap, err := readRelationship(ctx, cli, podKey)
			if err != nil {
				log.Errorf("readRelationship failed, err: %s", err.Error())
				return false
			}
			if chaosToContainerMap == nil {
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
				if _, ok := injectedContainers[name]; ok {
					return true
				}
			}
			return false
		},
	}
}

func buildChaosPredicate(ctx context.Context, cli client.Client, chaosKind string) predicate.Predicate {
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
				if (condition.Type == chaosmeshapi.ConditionSelected && condition.Status == corev1.ConditionFalse) ||
					(condition.Type == chaosmeshapi.ConditionAllInjected && condition.Status == corev1.ConditionFalse) {
					return false
				}
			}
			chaosKey := buildChaosKey(chaosKind, newObj)
			for _, record := range newStatus.Experiment.Records {
				podKey, containerName, err := parseRecordId(record.Id)
				if err != nil {
					log.Error(err, "failed to parse record id", "record id", record.Id, "chaos key", chaosKey)
					continue
				}
				chaosToContainerMap, err := readRelationship(ctx, cli, podKey)
				if err != nil {
					log.Errorf("readRelationship failed, err: %s", err.Error())
					continue
				}
				if chaosToContainerMap == nil {
					chaosToContainerMap = make(map[string]string)
					chaosToContainerMap[chaosKey] = containerName
				} else {
					chaosToContainerMap[chaosKey] = containerName
				}
				err = storeDataWithLock(ctx, cli, podKey, chaosToContainerMap)
				if err != nil {
					log.Errorf("storeDataWithLock failed, err: %s", err.Error())
				}
			}
			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			chaos := e.Object.(chaosmeshapi.StatefulObject)
			status := chaos.GetStatus()
			chaosKey := buildChaosKey(chaosKind, chaos)
			for _, record := range status.Experiment.Records {
				podKey, _, err := parseRecordId(record.Id)
				if err != nil {
					log.Error(err, "failed to parse record id", "record id", record.Id, "chaos key", chaosKey)
					continue
				}
				chaosToContainerMap, err := readRelationship(ctx, cli, podKey)
				if err != nil || chaosToContainerMap == nil {
					continue
				}
				delete(chaosToContainerMap, chaosKey)
				err = storeDataWithLock(ctx, cli, podKey, chaosToContainerMap)
				if err != nil {
					log.Errorf("storeDataWithLock failed, err: %s", err.Error())
				}
			}
			return false
		},
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

func parseRecordId(recordId string) (string, string, error) {
	parts := strings.Split(recordId, "/")
	if len(parts) == 3 {
		return fmt.Sprintf("%s/%s", parts[0], parts[1]), parts[2], nil
	}
	if len(parts) == 2 {
		return recordId, "", nil
	}
	return "", "", fmt.Errorf("invalid record id %s", recordId)
}

func readConfigFile(config *ControllerConfig, path string) error {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(contents, config)
}

func readRelationship(ctx context.Context, cli client.Client, podKey string) (map[string]string, error) {
	configMap := &corev1.ConfigMap{}
	namespacedName := types.NamespacedName{Name: podToChaosConfigMapName, Namespace: podToChaosConfigMapNamespace}

	if err := cli.Get(ctx, namespacedName, configMap); err != nil {
		return nil, err
	}

	podToChaos, err := deserializeData(configMap.Data)
	if err != nil {
		return nil, err
	}
	if chaosToContainerMap, ok := podToChaos[podKey]; ok {
		return chaosToContainerMap, nil
	}
	return nil, nil
}

func serializeData(data map[string]map[string]string) (map[string]string, error) {
	serializedValue, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return map[string]string{relationshipKey: string(serializedValue)}, nil
}

func deserializeData(rawData map[string]string) (map[string]map[string]string, error) {
	dataToBeDeserialized, ok := rawData[relationshipKey]
	if !ok {
		return nil, nil
	}

	data := make(map[string]map[string]string)
	err := json.Unmarshal([]byte(dataToBeDeserialized), &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func storeDataWithLock(ctx context.Context, cli client.Client, podKey string, data map[string]string) error {
	log.Infof("storeDataWithLock, podKey: %s, data: %v", podKey, data)
	namespacedName := types.NamespacedName{Name: podToChaosConfigMapName, Namespace: podToChaosConfigMapNamespace}
	maxRetries := 999
	for i := 0; i < maxRetries; i++ {
		configMap := &corev1.ConfigMap{}
		if err := cli.Get(ctx, namespacedName, configMap); err != nil {
			return err
		}
		existingData, err := deserializeData(configMap.Data)
		if err != nil {
			return err
		}
		mergeData := mergeMaps(data, existingData[podKey])
		if len(mergeData) == 0 {
			delete(existingData, podKey)
		} else {
			existingData[podKey] = mergeData
		}
		serializedData, err := serializeData(existingData)
		if err != nil {
			return err
		}

		configMap.Data = serializedData
		err = cli.Update(ctx, configMap)
		if err == nil {
			return nil
		}
		if errors.IsConflict(err) {
			backoffDuration := exponentialBackoff(i)
			log.Debugf("ResourceVersion conflict, retrying (%d/%d) after %v...\n", i+1, maxRetries, backoffDuration)
			time.Sleep(backoffDuration)
			continue
		}
		return err
	}

	return fmt.Errorf("failed to update ConfigMap after %d retries", maxRetries)
}

func exponentialBackoff(retry int) time.Duration {
	baseDelay := 100 * time.Millisecond
	maxDelay := 10 * time.Second

	delay := float64(baseDelay) * math.Pow(2, float64(retry))
	jitter := rand.Float64() * float64(baseDelay)
	delayWithJitter := time.Duration(delay + jitter)

	if delayWithJitter > maxDelay {
		return maxDelay
	}

	return delayWithJitter
}

// create podToChaosConfigMap if not exist,and with data: {"relationship":{}}
func createPodToChaosConfigMap(ctx context.Context, cli client.Client) error {
	configMap := &corev1.ConfigMap{}
	namespacedName := types.NamespacedName{Name: podToChaosConfigMapName, Namespace: podToChaosConfigMapNamespace}
	err := cli.Get(ctx, namespacedName, configMap)
	if err != nil && errors.IsNotFound(err) {
		if errors.IsNotFound(err) {
			configMap.Name = podToChaosConfigMapName
			configMap.Namespace = podToChaosConfigMapNamespace
			configMap.Data = map[string]string{relationshipKey: "{}"}
			err = cli.Create(ctx, configMap)
			if err != nil {
				return err
			}
			return nil
		}
		return err
	}
	return nil
}

func mergeMaps(dataToUpdate, latestDate map[string]string) map[string]string {
	mergedMap := make(map[string]string)
	for key, value := range dataToUpdate {
		mergedMap[key] = value
	}

	for key, value := range latestDate {
		mergedMap[key] = value
	}

	return mergedMap
}
