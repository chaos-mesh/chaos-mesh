// Copyright 2020 PingCAP, Inc.
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
	"fmt"
	"sort"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"

	"github.com/pingcap/chaos-mesh/test/e2e/config"
)

var ioTestConfigMap = `name: chaosfs-io
initContainers:
- name: inject-scripts
  image: %s
  imagePullPolicy: IfNotPresent
  command: ["sh", "-c", "/scripts/init.sh -d /var/run/data/test -f /var/run/data/fuse-data"]
containers:
- name: chaosfs
  image: %s
  imagePullPolicy: IfNotPresent
  ports:
  - containerPort: 65534
  securityContext:
    privileged: true
  command:
    - /usr/local/bin/chaosfs
    - -addr=:65534
    - -pidfile=/tmp/fuse/pid
    - -original=/var/run/data/fuse-data
    - -mountpoint=/var/run/data/test
  volumeMounts:
    - name: datadir
      mountPath: /var/run/data
      mountPropagation: Bidirectional
volumeMounts:
  - name: datadir
    mountPath: /var/run/data
    mountPropagation: HostToContainer
  - name: scripts
    mountPath: /tmp/scripts
  - name: fuse
    mountPath: /tmp/fuse
volumes:
  - name: scripts
    emptyDir: {}
  - name: fuse
    emptyDir: {}
postStart:
  io:
    command:
      - /tmp/scripts/wait-fuse.sh
`

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
				"app": "timer",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "timer",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "timer",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image:           config.TestConfig.E2EImage,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Name:            "timer",
							Command:         []string{"/bin/test"},
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
						"admission-webhook.pingcap.com/request": "chaosfs-io",
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

func NewIOTestConfigMap(name, namespace string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
			Labels: map[string]string{
				"app.kubernetes.io/component": "webhook",
			},
		},
		Data: map[string]string{
			"chaosfs-io.yaml": fmt.Sprintf(ioTestConfigMap,
				config.TestConfig.ChaosScriptsImage, config.TestConfig.ChaosFSImage),
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
