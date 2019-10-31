package inject

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/golang/glog"
	"github.com/pingcap/chaos-operator/pkg/webhook/config"

	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ignoredNamespaces = []string{
	metav1.NamespaceSystem,
	metav1.NamespacePublic,
}

const (
	// StatusInjected is the annotation value for /status that indicates an injection was already performed on this pod
	StatusInjected = "injected"
)

func Inject(res *v1beta1.AdmissionRequest, cfg *config.Config) *v1beta1.AdmissionResponse {
	var pod corev1.Pod
	if err := json.Unmarshal(res.Object.Raw, &pod); err != nil {
		glog.Errorf("Could not unmarshal raw object: %v", err)
		return &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	glog.Infof("AdmissionReview for Kind=%s, Namespace=%s Name=%s (%s) UID=%s patchOperation=%s UserInfo=%s",
		res.Kind, res.Namespace, res.Name, pod.Name, res.UID, res.Operation, res.UserInfo)

	requiredKey, ok := injectRequired(ignoredNamespaces, &pod.ObjectMeta, cfg)
	if !ok {
		glog.Infof("Skipping injection for %s/%s due to policy check", pod.Namespace, pod.Name)
		return &v1beta1.AdmissionResponse{
			Allowed: true,
		}
	}

	injectionConfig, err := cfg.GetRequestedConfig(requiredKey)
	if err != nil {
		glog.Errorf("Error getting injection config %v, permitting launch of pod with no sidecar injected: %s",
			injectionConfig, err)
		// dont prevent pods from launching! just return allowed
		return &v1beta1.AdmissionResponse{
			Allowed: true,
		}
	}

	annotations := map[string]string{}
	annotations[cfg.StatusAnnotationKey()] = StatusInjected
	patchBytes, err := createPatch(&pod, injectionConfig, annotations)
	if err != nil {
		return &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	glog.Infof("AdmissionResponse: patch=%v\n", string(patchBytes))
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
func injectRequired(ignoredList []string, metadata *metav1.ObjectMeta, cfg *config.Config) (string, bool) {
	// skip special kubernetes system namespaces
	for _, namespace := range ignoredList {
		if metadata.Namespace == namespace {
			glog.Infof("Skip mutation for %v for it' in special namespace:%v", metadata.Name, metadata.Namespace)
			return "", false
		}
	}

	annotations := metadata.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	status, ok := annotations[cfg.StatusAnnotationKey()]
	if ok && strings.ToLower(status) == StatusInjected {
		glog.Infof("Pod %s/%s annotation %s=%s indicates injection already satisfied, skipping",
			metadata.Namespace, metadata.Name, cfg.StatusAnnotationKey(), status)
		return "", false
	}

	required, ok := annotations[cfg.RequestAnnotationKey()]
	if !ok {
		glog.Infof("Pod %s/%s annotation %s is missing, skipping injection",
			metadata.Namespace, metadata.Name, cfg.RequestAnnotationKey())
		return "", false
	}

	requiredConfig := strings.ToLower(required)
	glog.Infof("Pod %s/%s annotation %s=%s requesting sidecar config %s",
		metadata.Namespace, metadata.Name, cfg.RequestAnnotationKey(), required, requiredConfig)
	return requiredConfig, true
}

// create mutation patch for resource
func createPatch(pod *corev1.Pod, inj *config.InjectionConfig, annotations map[string]string) ([]byte, error) {
	var patch []patchOperation

	mutatedInjectedContainers := mergeEnvVars(inj.Environment, inj.Containers)
	mutatedInjectedContainers = mergeVolumeMounts(inj.VolumeMounts, mutatedInjectedContainers)

	mutatedInjectedInitContainers := mergeEnvVars(inj.Environment, inj.InitContainers)
	mutatedInjectedInitContainers = mergeVolumeMounts(inj.VolumeMounts, mutatedInjectedInitContainers)

	patch = append(patch, addContainers(pod.Spec.Containers, mutatedInjectedContainers, "/spec/containers")...)

	patch = append(patch, setEnvironment(pod.Spec.Containers, inj.Environment)...)
	patch = append(patch, addVolumeMounts(pod.Spec.Containers, inj.VolumeMounts)...)

	patch = append(patch, addContainers(pod.Spec.InitContainers, mutatedInjectedInitContainers, "/spec/initContainers")...)
	patch = append(patch, addHostAliases(pod.Spec.HostAliases, inj.HostAliases, "/spec/hostAliases")...)
	patch = append(patch, addVolumes(pod.Spec.Volumes, inj.Volumes, "/spec/volumes")...)

	patch = append(patch, updateAnnotations(pod.Annotations, annotations)...)

	patch = append(patch, updatePIDShare(inj.ShareProcessNamespace)...)

	return json.Marshal(patch)
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

func addVolumeMounts(target []corev1.Container, addedVolumeMounts []corev1.VolumeMount) (patch []patchOperation) {
	var value interface{}
	for containerIndex, container := range target {
		// for each container in the spec, determine if we want to patch with any volume mounts
		first := len(container.VolumeMounts) == 0
		for _, add := range addedVolumeMounts {
			path := fmt.Sprintf("/spec/containers/%d/volumeMounts", containerIndex)
			hasKey := false
			// make sure we dont override any existing volume mounts; we only add, dont replace
			for _, origVolumeMount := range container.VolumeMounts {
				if origVolumeMount.Name == add.Name {
					hasKey = true
					break
				}
			}
			if !hasKey {
				// make a patch
				value = add
				if first {
					first = false
					value = []corev1.VolumeMount{add}
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
		keyEscaped := strings.Replace(key, "/", "~1", -1)

		if target == nil || target[key] == "" {
			target = map[string]string{}
			patch = append(patch, patchOperation{
				Op:    "add",
				Path:  "/metadata/annotations/" + keyEscaped,
				Value: value,
			})
		} else {
			patch = append(patch, patchOperation{
				Op:    "replace",
				Path:  "/metadata/annotations/" + keyEscaped,
				Value: value,
			})
		}
	}
	return patch
}

func updatePIDShare(value bool) (patch []patchOperation) {

	op := "add"
	patch = append(patch, patchOperation{
		Op:    op,
		Path:  "/spec/shareProcessNamespace",
		Value: value,
	})
	return patch
}
