// Copyright 2020 Chaos Mesh Authors.
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

package fixture

import (
	"sort"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"

	"github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/config"
)

// NewCommonNginxPod describe that we use common nginx pod to be tested in our chaos-operator test
func NewCommonNginxPod(name, namespace string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app": "nginx",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Image:           "nginx:latest",
					ImagePullPolicy: corev1.PullIfNotPresent,
					Name:            "nginx",
				},
			},
		},
	}
}

// NewCommonNginxDeployment would create a nginx deployment
func NewCommonNginxDeployment(name, namespace string, replicas int32) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app": "nginx",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "nginx",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "nginx",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image:           "nginx:latest",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Name:            "nginx",
						},
					},
				},
			},
		},
	}
}

// NewTimerDeployment creates a timer deployment
func NewTimerDeployment(name, namespace string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image:           config.TestConfig.E2EImage,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Name:            name,
							Command:         []string{"/bin/test"},
						},
					},
				},
			},
		},
	}
}

// NewNetworkTestDeployment creates a deployment for e2e test
func NewNetworkTestDeployment(name, namespace string, extraLabels map[string]string) *appsv1.Deployment {
	labels := map[string]string{
		"app": name,
	}
	for key, val := range extraLabels {
		labels[key] = val
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image:           config.TestConfig.E2EImage,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Name:            "network",
							Command:         []string{"/bin/test"},
						},
					},
				},
			},
		},
	}
}

// NewStressTestDeployment creates a deployment for e2e test
func NewStressTestDeployment(name, namespace string, extraLabels map[string]string) *appsv1.Deployment {
	labels := map[string]string{
		"app": name,
	}
	for key, val := range extraLabels {
		labels[key] = val
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image:           config.TestConfig.E2EImage,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Name:            "stress",
							Command:         []string{"/bin/test"},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("0"),
									corev1.ResourceMemory: resource.MustParse("0"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1"),
									corev1.ResourceMemory: resource.MustParse("150M"),
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "sys",
									MountPath: "/sys",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "sys",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/sys",
								},
							},
						},
					},
				},
			},
		},
	}
}

// NewIOTestDeployment creates a deployment for e2e test
func NewIOTestDeployment(name, namespace string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app": "io",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "io",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "io",
					},
					Annotations: map[string]string{
						"admission-webhook.chaos-mesh.org/request": "chaosfs-io",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image:           config.TestConfig.E2EImage,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Name:            "io",
							Command:         []string{"/bin/test"},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "datadir",
									MountPath: "/var/run/data",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "datadir",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}
}

// NewHTTPTestDeployment creates a deployment for e2e test
func NewHTTPTestDeployment(name, namespace string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app": "http",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "http",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "http",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image:           config.TestConfig.E2EImage,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Name:            "http",
							Command:         []string{"/bin/test"},
						},
					},
				},
			},
		},
	}
}

// NewE2EService creates a service for the E2E helper deployment
func NewE2EService(name, namespace string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeNodePort,
			Selector: map[string]string{
				"app": name,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       8080,
					TargetPort: intstr.IntOrString{IntVal: 8080},
				},
				// Only used in network chaos
				{
					Name:       "nc-port",
					Port:       1070,
					TargetPort: intstr.IntOrString{IntVal: 8000},
				},
				// Only used in io chaos
				{
					Name:       "chaosfs",
					Port:       65534,
					TargetPort: intstr.IntOrString{IntVal: 65534},
				},
			},
		},
	}
}

// HaveSameUIDs returns if pods1 and pods2 are same based on their UIDs
func HaveSameUIDs(pods1 []corev1.Pod, pods2 []corev1.Pod) bool {
	count := len(pods1)
	if count != len(pods2) {
		return false
	}
	ids1, ids2 := make([]string, count), make([]string, count)
	for i := 0; i < count; i++ {
		ids1[i], ids2[i] = string(pods1[i].UID), string(pods2[i].UID)
	}
	sort.Strings(ids1)
	sort.Strings(ids2)
	for i := 0; i < count; i++ {
		if ids1[i] != ids2[i] {
			return false
		}
	}
	return true
}
