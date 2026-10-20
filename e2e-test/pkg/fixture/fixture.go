// Copyright 2021 Chaos Mesh Authors.
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

package fixture

import (
	"fmt"
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
			Replicas: pointer.Int32(1),
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
			Replicas: pointer.Int32(1),
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
			Replicas: pointer.Int32(1),
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
			Replicas: pointer.Int32(1),
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
			Replicas: pointer.Int32(1),
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
				{
					Name:       "https",
					Port:       8081,
					TargetPort: intstr.IntOrString{IntVal: 8081},
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

// NewJavaHelloWorldPod creates a pod running a small java program.
// The program calls Main.sayhello once per second, and sayhello prints one
// log line. It is used by JVMChaos e2e tests, which inject faults into the
// sayhello method. The image uses JDK 21 on purpose: the byteman rule for
// the latency action must work with the Thread.sleep(Duration) overload
// added in JDK 19, see https://github.com/chaos-mesh/chaos-mesh/pull/4821.
func NewJavaHelloWorldPod(name, namespace string) *corev1.Pod {
	javaSource := `
public class Main {
    public static void main(String[] args) throws Exception {
        for (long i = 0; ; i++) {
            try {
                sayhello(i);
            } catch (Exception e) {
                System.out.println("Got an exception! " + e);
            }
            Thread.sleep(1000);
        }
    }

    public static void sayhello(long num) throws Exception {
        System.out.println(getnum(num) + ". Hello World");
    }

    public static long getnum(long num) {
        return num;
    }
}
`
	script := fmt.Sprintf(`set -e
mkdir -p /app
cat > /app/Main.java <<'EOF'
%s
EOF
cd /app
javac Main.java
exec java Main
`, javaSource)

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:            name,
					Image:           "eclipse-temurin:21-jdk",
					ImagePullPolicy: corev1.PullIfNotPresent,
					Command:         []string{"sh", "-c", script},
				},
			},
		},
	}
}

// NewMySQLDeployment creates a single-instance MySQL deployment used as
// the backend for JVMChaos mysql action tests.
func NewMySQLDeployment(namespace string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mysql",
			Namespace: namespace,
			Labels: map[string]string{
				"app": "mysql",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "mysql",
				},
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "mysql",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "mysql",
							Image:           "mysql:8.4",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env: []corev1.EnvVar{
								{
									Name:  "MYSQL_ALLOW_EMPTY_PASSWORD",
									Value: "yes",
								},
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 3306,
								},
							},
						},
					},
				},
			},
		},
	}
}

// NewMySQLService exposes the MySQL deployment inside the cluster.
func NewMySQLService(namespace string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mysql",
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "mysql",
			},
			Ports: []corev1.ServicePort{
				{
					Port:       3306,
					TargetPort: intstr.FromInt(3306),
				},
			},
		},
	}
}

// NewMySQLQueryDeployment creates the mysql query java application. It
// exposes an HTTP endpoint /query?sql=... and runs the SQL against the
// DSN configured through the environment.
func NewMySQLQueryDeployment(namespace, dsn string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mysql-query",
			Namespace: namespace,
			Labels: map[string]string{
				"app": "mysql-query",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "mysql-query",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "mysql-query",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "mysql-query",
							Image:           "xiang13225080/mysqldemo:latest",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env: []corev1.EnvVar{
								{
									Name:  "MYSQL_DSN",
									Value: dsn,
								},
								{
									Name:  "MYSQL_USER",
									Value: "root",
								},
								{
									Name:  "MYSQL_PASSWORD",
									Value: "",
								},
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8001,
								},
							},
						},
					},
				},
			},
		},
	}
}

// NewMySQLQueryService exposes the mysql query application inside the cluster.
func NewMySQLQueryService(namespace string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mysql-query",
			Namespace: namespace,
			Labels: map[string]string{
				"app": "mysql-query",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "mysql-query",
			},
			Ports: []corev1.ServicePort{
				{
					Port:       8001,
					TargetPort: intstr.FromInt(8001),
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
