// Copyright 2019 Chaos Mesh Authors.
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

package inject

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/chaos-mesh/chaos-mesh/controllers/metrics"
	"github.com/chaos-mesh/chaos-mesh/pkg/annotation"
	controllerCfg "github.com/chaos-mesh/chaos-mesh/pkg/config"
	podselector "github.com/chaos-mesh/chaos-mesh/pkg/selector/pod"
	"github.com/chaos-mesh/chaos-mesh/pkg/webhook/config"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var log = ctrl.Log.WithName("inject-webhook")

var ignoredNamespaces = []string{
	metav1.NamespaceSystem,
	metav1.NamespacePublic,
}

const (
	// StatusInjected is the annotation value for /status that indicates an injection was already performed on this pod
	StatusInjected = "injected"
)

// Inject do pod template config inject
func Inject(res *v1beta1.AdmissionRequest, cli client.Client, cfg *config.Config, controllerCfg *controllerCfg.ChaosControllerConfig, metrics *metrics.ChaosCollector) *v1beta1.AdmissionResponse {
	var pod corev1.Pod
	if err := json.Unmarshal(res.Object.Raw, &pod); err != nil {
		log.Error(err, "Could not unmarshal raw object")
		return &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	// Deal with potential empty fields, e.g., when the pod is created by a deployment
	podName := potentialPodName(&pod.ObjectMeta)
	if pod.ObjectMeta.Namespace == "" {
		pod.ObjectMeta.Namespace = res.Namespace
	}

	log.Info("AdmissionReview for",
		"Kind", res.Kind, "Namespace", res.Namespace, "Name", res.Name, "podName", podName, "UID", res.UID, "patchOperation", res.Operation, "UserInfo", res.UserInfo)
	log.V(4).Info("Object", "Object", string(res.Object.Raw))
	log.V(4).Info("OldObject", "OldObject", string(res.OldObject.Raw))
	log.V(4).Info("Pod", "Pod", pod)

	requiredKey, ok := injectRequired(&pod.ObjectMeta, cli, cfg, controllerCfg)
	if !ok {
		log.Info("Skipping injection due to policy check", "namespace", pod.ObjectMeta.Namespace, "name", podName)
		return &v1beta1.AdmissionResponse{
			Allowed: true,
		}
	}

	if metrics != nil {
		metrics.InjectRequired.WithLabelValues(res.Namespace, requiredKey).Inc()
	}
	injectionConfig, err := cfg.GetRequestedConfig(pod.Namespace, requiredKey)
	if err != nil {
		log.Error(err, "Error getting injection config, permitting launch of pod with no sidecar injected", "injectionConfig",
			injectionConfig)
		// dont prevent pods from launching! just return allowed
		return &v1beta1.AdmissionResponse{
			Allowed: true,
		}
	}

	if injectionConfig.Selector != nil {
		meet, err := podselector.CheckPodMeetSelector(pod, *injectionConfig.Selector)
		if err != nil {
			log.Error(err, "Failed to check pod selector", "namespace", pod.Namespace)
			return &v1beta1.AdmissionResponse{
				Allowed: true,
			}
		}

		if !meet {
			log.Info("Skipping injection, this pod does not meet the selection criteria",
				"namespace", pod.Namespace, "name", pod.Name)
			return &v1beta1.AdmissionResponse{
				Allowed: true,
			}
		}
	}

	annotations := map[string]string{cfg.StatusAnnotationKey(): StatusInjected}

	patchBytes, err := createPatch(&pod, injectionConfig, annotations)
	if err != nil {
		return &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	log.Info("AdmissionResponse: patch", "patchBytes", string(patchBytes))
	if metrics != nil {
		metrics.Injections.WithLabelValues(res.Namespace, requiredKey).Inc()
	}
	return &v1beta1.AdmissionResponse{
		Allowed: true,
		Patch:   patchBytes,
		PatchType: func() *v1beta1.PatchType {
			pt := v1beta1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}

// Check whether the target resource need to be injected and return the required config name
func injectRequired(metadata *metav1.ObjectMeta, cli client.Client, cfg *config.Config, controllerCfg *controllerCfg.ChaosControllerConfig) (string, bool) {
	// skip special kubernetes system namespaces
	for _, namespace := range ignoredNamespaces {
		if metadata.Namespace == namespace {
			log.Info("Skip mutation for it' in special namespace", "name", metadata.Name, "namespace", metadata.Namespace)
			return "", false
		}
	}

	if controllerCfg.EnableFilterNamespace {
		ok, err := podselector.IsAllowedNamespaces(context.Background(), cli, metadata.Namespace)
		if err != nil {
			log.Error(err, "fail to check whether this namespace should be injected", "namespace", metadata.Namespace)
		}

		if !ok {
			log.Info("Skip mutation for it' in special namespace", "name", metadata.Name, "namespace", metadata.Namespace)
			return "", false
		}
	}

	log.V(4).Info("meta", "meta", metadata)

	if checkInjectStatus(metadata, cfg) {
		log.Info("Pod annotation indicates injection already satisfied, skipping",
			"namespace", metadata.Namespace, "name", metadata.Name,
			"annotationKey", cfg.StatusAnnotationKey(), "value", StatusInjected)
		return "", false
	}

	requiredConfig, ok := injectByPodRequired(metadata, cfg)
	if ok {
		log.Info("Pod annotation requesting sidecar config",
			"namespace", metadata.Namespace, "name", metadata.Name,
			"annotation", cfg.RequestAnnotationKey(), "requiredConfig", requiredConfig)
		return requiredConfig, true
	}

	requiredConfig, ok = injectByNamespaceRequired(metadata, cli, cfg)
	if ok {
		log.Info("Pod annotation requesting sidecar config",
			"namespace", metadata.Namespace, "name", metadata.Name,
			"annotation", cfg.RequestAnnotationKey(), "requiredConfig", requiredConfig)
		return requiredConfig, true
	}

	requiredConfig, ok = injectByNamespaceInitRequired(metadata, cli, cfg)
	if ok {
		log.Info("Pod annotation init requesting sidecar config",
			"namespace", metadata.Namespace, "name", metadata.Name,
			"annotation", cfg.RequestAnnotationKey(), "requiredConfig", requiredConfig)
		return requiredConfig, true
	}

	return "", false
}

func checkInjectStatus(metadata *metav1.ObjectMeta, cfg *config.Config) bool {
	annotations := metadata.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	status, ok := annotations[cfg.StatusAnnotationKey()]
	if ok && strings.ToLower(status) == StatusInjected {
		return true
	}

	return false
}

func injectByNamespaceRequired(metadata *metav1.ObjectMeta, cli client.Client, cfg *config.Config) (string, bool) {
	var ns corev1.Namespace
	if err := cli.Get(context.Background(), types.NamespacedName{Name: metadata.Namespace}, &ns); err != nil {
		log.Error(err, "failed to get namespace", "namespace", metadata.Namespace)
		return "", false
	}
	annotations := ns.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	required, ok := annotations[annotation.GenKeyForWebhook(cfg.RequestAnnotationKey(), metadata.Name)]
	if !ok {
		log.Info("Pod annotation by namespace is missing, skipping injection",
			"namespace", metadata.Namespace, "pod", metadata.Name, "config", required)
		return "", false
	}

	log.Info("Get sidecar config from namespace annotations",
		"namespace", metadata.Namespace, "pod", metadata.Name, "config", required)
	return strings.ToLower(required), true
}

func injectByNamespaceInitRequired(metadata *metav1.ObjectMeta, cli client.Client, cfg *config.Config) (string, bool) {
	var ns corev1.Namespace
	if err := cli.Get(context.Background(), types.NamespacedName{Name: metadata.Namespace}, &ns); err != nil {
		log.Error(err, "failed to get namespace", "namespace", metadata.Namespace)
		return "", false
	}

	annotations := ns.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	required, ok := annotations[cfg.RequestInitAnnotationKey()]
	if !ok {
		log.Info("Pod annotation by namespace is missing, skipping injection",
			"namespace", metadata.Namespace, "pod", metadata.Name, "config", required)
		return "", false
	}

	log.Info("Get sidecar config from namespace annotations",
		"namespace", metadata.Namespace, "pod", metadata.Name, "config", required)
	return strings.ToLower(required), true
}

func injectByPodRequired(metadata *metav1.ObjectMeta, cfg *config.Config) (string, bool) {
	annotations := metadata.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	required, ok := annotations[cfg.RequestAnnotationKey()]
	if !ok {
		log.Info("Pod annotation is missing, skipping injection",
			"namespace", metadata.Namespace, "name", metadata.Name, "annotation", cfg.RequestAnnotationKey())
		return "", false
	}

	log.Info("Get sidecar config from pod annotations",
		"namespace", metadata.Namespace, "pod", metadata.Name, "config", required)
	return strings.ToLower(required), true
}

// create mutation patch for resource
func createPatch(pod *corev1.Pod, inj *config.InjectionConfig, annotations map[string]string) ([]byte, error) {
	var patch []patchOperation

	// make sure any injected containers in our config get the EnvVars and VolumeMounts injected
	// this mutates inj.Containers with our environment vars
	mutatedInjectedContainers := mergeEnvVars(inj.Environment, inj.Containers)
	mutatedInjectedContainers = mergeVolumeMounts(inj.VolumeMounts, mutatedInjectedContainers)

	// make sure any injected init containers in our config get the EnvVars and VolumeMounts injected
	// this mutates inj.InitContainers with our environment vars
	mutatedInjectedInitContainers := mergeEnvVars(inj.Environment, inj.InitContainers)
	mutatedInjectedInitContainers = mergeVolumeMounts(inj.VolumeMounts, mutatedInjectedInitContainers)

	// patch all existing containers with the env vars and volume mounts
	patch = append(patch, setVolumeMounts(pod.Spec.Containers, inj.VolumeMounts, "/spec/containers")...)
	// TODO: fix set env
	// setEnvironment may not work, because we replace the whole container in `setVolumeMounts`
	patch = append(patch, setEnvironment(pod.Spec.Containers, inj.Environment)...)

	// patch containers with our injected containers
	patch = append(patch, addContainers(pod.Spec.Containers, mutatedInjectedContainers, "/spec/containers")...)

	// add initContainers, hostAliases and volumes
	patch = append(patch, addContainers(pod.Spec.InitContainers, mutatedInjectedInitContainers, "/spec/initContainers")...)
	patch = append(patch, addHostAliases(pod.Spec.HostAliases, inj.HostAliases, "/spec/hostAliases")...)
	patch = append(patch, addVolumes(pod.Spec.Volumes, inj.Volumes, "/spec/volumes")...)

	// set annotations
	patch = append(patch, updateAnnotations(pod.Annotations, annotations)...)

	// set shareProcessNamespace
	patch = append(patch, updateShareProcessNamespace(inj.ShareProcessNamespace)...)

	// TODO: remove injecting commands when sidecar container supported
	// set commands and args
	patch = append(patch, setCommands(pod.Spec.Containers, inj.PostStart)...)

	return json.Marshal(patch)
}

func setCommands(target []corev1.Container, postStart map[string]config.ExecAction) (patch []patchOperation) {
	if postStart == nil {
		return
	}

	for containerIndex, container := range target {
		execCmd, ok := postStart[container.Name]
		if !ok {
			continue
		}

		path := fmt.Sprintf("/spec/containers/%d/command", containerIndex)

		commands := MergeCommands(execCmd.Command, container.Command, container.Args)

		log.Info("Inject command", "command", commands)

		patch = append(patch, patchOperation{
			Op:    "replace",
			Path:  path,
			Value: commands,
		})

		argsPath := fmt.Sprintf("/spec/containers/%d/args", containerIndex)
		patch = append(patch, patchOperation{
			Op:    "replace",
			Path:  argsPath,
			Value: []string{},
		})
	}
	return patch
}

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func setEnvironment(target []corev1.Container, addedEnv []corev1.EnvVar) (patch []patchOperation) {
	var value interface{}
	for containerIndex, container := range target {
		// for each container in the spec, determine if we want to patch with any env vars
		first := len(container.Env) == 0
		for _, add := range addedEnv {
			path := fmt.Sprintf("/spec/containers/%d/env", containerIndex)
			hasKey := false
			// make sure we dont override any existing env vars; we only add, dont replace
			for _, origEnv := range container.Env {
				if origEnv.Name == add.Name {
					hasKey = true
					break
				}
			}
			if !hasKey {
				// make a patch
				value = add
				if first {
					first = false
					value = []corev1.EnvVar{add}
				} else {
					path = path + "/-"
				}
				patch = append(patch, patchOperation{
					Op:    "add",
					Path:  path,
					Value: value,
				})
			}
		}
	}

	return patch
}

func addContainers(target, added []corev1.Container, basePath string) (patch []patchOperation) {
	first := len(target) == 0
	var value interface{}
	for _, add := range added {
		value = add
		log.V(6).Info("Add container", "add", add)
		path := basePath
		if first {
			first = false
			value = []corev1.Container{add}
		} else {
			path = path + "/-"
		}
		patch = append(patch, patchOperation{
			Op:    "add",
			Path:  path,
			Value: value,
		})
	}
	return patch
}

func addVolumes(target, added []corev1.Volume, basePath string) (patch []patchOperation) {
	first := len(target) == 0
	var value interface{}
	for _, add := range added {
		value = add
		path := basePath
		if first {
			first = false
			value = []corev1.Volume{add}
		} else {
			path = path + "/-"
		}
		patch = append(patch, patchOperation{
			Op:    "add",
			Path:  path,
			Value: value,
		})
	}
	return patch
}

func setVolumeMounts(target []corev1.Container, addedVolumeMounts []corev1.VolumeMount, basePath string) (patch []patchOperation) {
	for index, c := range target {
		volumeMounts := map[string]corev1.VolumeMount{}
		for _, vm := range c.VolumeMounts {
			volumeMounts[vm.Name] = vm
		}
		for _, added := range addedVolumeMounts {
			log.Info("volumeMount", "add", added)
			volumeMounts[added.Name] = added
		}

		vs := []corev1.VolumeMount{}
		for _, vm := range volumeMounts {
			vs = append(vs, vm)
		}
		target[index].VolumeMounts = vs
	}

	patch = append(patch, patchOperation{
		Op:    "replace",
		Path:  basePath,
		Value: target,
	})

	return patch
}

func addHostAliases(target, added []corev1.HostAlias, basePath string) (patch []patchOperation) {
	first := len(target) == 0
	var value interface{}
	for _, add := range added {
		value = add
		path := basePath
		if first {
			first = false
			value = []corev1.HostAlias{add}
		} else {
			path = path + "/-"
		}
		patch = append(patch, patchOperation{
			Op:    "add",
			Path:  path,
			Value: value,
		})
	}
	return patch
}

// for containers, add any env vars that are not already defined in the Env list.
// this does _not_ return patches; this is intended to be used only on containers defined
// in the injection config, so the resources do not exist yet in the k8s api (thus no patch needed)
func mergeEnvVars(envs []corev1.EnvVar, containers []corev1.Container) []corev1.Container {
	mutatedContainers := []corev1.Container{}
	for _, c := range containers {
		for _, newEnv := range envs {
			// check each container for each env var by name.
			// if the container has a matching name, dont override!
			skip := false
			for _, origEnv := range c.Env {
				if origEnv.Name == newEnv.Name {
					skip = true
					break
				}
			}
			if !skip {
				c.Env = append(c.Env, newEnv)
			}
		}
		mutatedContainers = append(mutatedContainers, c)
	}
	return mutatedContainers
}

func mergeVolumeMounts(volumeMounts []corev1.VolumeMount, containers []corev1.Container) []corev1.Container {
	mutatedContainers := []corev1.Container{}
	for _, c := range containers {
		for _, newVolumeMount := range volumeMounts {
			// check each container for each volume mount by name.
			// if the container has a matching name, dont override!
			skip := false
			for _, origVolumeMount := range c.VolumeMounts {
				if origVolumeMount.Name == newVolumeMount.Name {
					skip = true
					break
				}
			}
			if !skip {
				c.VolumeMounts = append(c.VolumeMounts, newVolumeMount)
			}
		}
		mutatedContainers = append(mutatedContainers, c)
	}
	return mutatedContainers
}

func updateAnnotations(target map[string]string, added map[string]string) (patch []patchOperation) {
	for key, value := range added {
		if target == nil || target[key] == "" {
			target = map[string]string{}
			patch = append(patch, patchOperation{
				Op:   "add",
				Path: "/metadata/annotations",
				Value: map[string]string{
					key: value,
				},
			})
		} else {
			patch = append(patch, patchOperation{
				Op:    "replace",
				Path:  "/metadata/annotations/" + key,
				Value: value,
			})
		}
	}
	return patch
}

func updateShareProcessNamespace(value bool) (patch []patchOperation) {
	op := "add"
	patch = append(patch, patchOperation{
		Op:    op,
		Path:  "/spec/shareProcessNamespace",
		Value: value,
	})
	return patch
}

func potentialPodName(metadata *metav1.ObjectMeta) string {
	if metadata.Name != "" {
		return metadata.Name
	}
	if metadata.GenerateName != "" {
		return metadata.GenerateName + "***** (actual name not yet known)"
	}
	return ""
}
