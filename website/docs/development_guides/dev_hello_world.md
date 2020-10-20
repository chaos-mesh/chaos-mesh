---
id: develop_a_new_chaos 
title: Develop a New Chaos
sidebar_label: Develop a New Chaos
---

After [preparing the development environment](setup_env.md), let's develop a new type of chaos, HelloWorldChaos, which only prints a "Hello World!" message to the log. Generally, to add a new chaos type for Chaos Mesh, you need to take the following steps:

1. [Define the schema type](#define-the-schema-type)
2. [Register the CRD](#register-the-crd)
3. [Register the handler for this chaos object](#register-the-handler-for-this-chaos-object)
4. [Make the Docker image](#make-the-docker-image)
5. [Run chaos](#run-chaos)

## Define the schema type

To define the schema type for the new chaos object, add `helloworldchaos_types.go` in the api directory [`/api/v1alpha1`](https://github.com/chaos-mesh/chaos-mesh/tree/master/api/v1alpha1) and fill it with the following content:

```go
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +chaos-mesh:base

// HelloWorldChaos is the Schema for the helloworldchaos API
type HelloWorldChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HelloWorldChaosSpec   `json:"spec"`
	Status HelloWorldChaosStatus `json:"status,omitempty"`
}

// HelloWorldChaosSpec is the content of the specification for a HelloWorldChaos
type HelloWorldChaosSpec struct {
	// Duration represents the duration of the chaos action
	// +optional
	Duration *string `json:"duration,omitempty"`

	// Scheduler defines some schedule rules to control the running time of the chaos experiment about time.
	// +optional
	Scheduler *SchedulerSpec `json:"scheduler,omitempty"`
}

// HelloWorldChaosStatus represents the status of a HelloWorldChaos
type HelloWorldChaosStatus struct {
	ChaosStatus `json:",inline"`
}
```

With this file added, the HelloWorldChaos schema type is defined. The structure of it can be described as the YAML file below:

```yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: HelloWorldChaos
metadata:
  name: <name-of-this-resource>
  namespace: <ns-of-this-resource>
spec:
  duration: <duration-of-every-action>
  scheduler:
    cron: <the-cron-job-definition-of-this-chaos>
status:
  phase: <phase-of-this-resource>
  ...
```

`make generate` will generate boilerplate functions for it, which is needed to integrate the resource in the Chaos Mesh.

## Register the CRD

The HelloWorldChaos object is a custom resource object in Kubernetes. This means you need to register the corresponding CRD in the Kubernetes API. Run `make yaml`, then the CRD will be generated in `/config/crd/bases/chaos-mesh.org_helloworldchaos.yaml`. In order to combine all these YAML file into `/manifests/crd.yaml`, modify [kustomization.yaml](https://github.com/chaos-mesh/chaos-mesh/blob/master/config/crd/kustomization.yaml) by adding the corresponding line as shown below:

```yaml
resources:
- bases/chaos-mesh.org_podchaos.yaml
- bases/chaos-mesh.org_networkchaos.yaml
- bases/chaos-mesh.org_iochaos.yaml
- bases/chaos-mesh.org_helloworldchaos.yaml  # this is the new line
```

Then the definition of HelloWorldChaos will show in `/manifests/crd.yaml`. You can check it through `git diff`

## Register the handler for this chaos object

Create file `/controllers/helloworldchaos/endpoint.go` and fill it with following codes:

```go
package helloworldchaos

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/router"
	ctx "github.com/chaos-mesh/chaos-mesh/pkg/router/context"
	end "github.com/chaos-mesh/chaos-mesh/pkg/router/endpoint"
)

type endpoint struct {
	ctx.Context
}

func (e *endpoint) Apply(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	e.Log.Info("Hello World!")
	return nil
}

func (e *endpoint) Recover(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	return nil
}

func (e *endpoint) Object() v1alpha1.InnerObject {
	return &v1alpha1.HelloWorldChaos{}
}

func init() {
	router.Register("helloworldchaos", &v1alpha1.HelloWorldChaos{}, func(obj runtime.Object) bool {
		return true
	}, func(ctx ctx.Context) end.Endpoint {
		return &endpoint{
			Context: ctx,
		}
	})
}
```

We should also import `github.com/chaos-mesh/chaos-mesh/controllers/helloworldchaos` in the `/cmd/controller-manager/main.go`, then it will register on the route table when the controller starts up.

## Make the Docker image

Having the object successfully added, you can make a Docker image and push it to your registry:

```bash
make
make docker-push
```

> **Note:**
>
> The default `DOCKER_REGISTRY` is `localhost:5000`, which is preset in `hack/kind-cluster-build.sh`. You can overwrite it to any registry to which you have access permission.

## Run chaos

You are almost there. In this step, you will pull the image and apply it for testing.

Before you pull any image for Chaos Mesh (using `helm install` or `helm upgrade`), modify [values.yaml](https://github.com/chaos-mesh/chaos-mesh/blob/master/helm/chaos-mesh/values.yaml) of helm template to replace the default image with what you just pushed to your local registry.

In this case, the template uses `pingcap/chaos-mesh:latest` as the default target registry, so you need to replace it with `localhost:5000`, as shown below:

```yaml
clusterScoped: true

# Also see clusterScoped and controllerManager.serviceAccount
rbac:
  create: true

controllerManager:
  serviceAccount: chaos-controller-manager
  ...
  image: localhost:5000/pingcap/chaos-mesh:latest
  ...
chaosDaemon:
  image: localhost:5000/pingcap/chaos-daemon:latest
  ...
dashboard:
  image: localhost:5000/pingcap/chaos-dashboard:latest
  ...
```

Now take the following steps to run chaos:

1. Get the related custom resource type for Chaos Mesh:

    ```bash
    kubectl apply -f manifests/
    kubectl get crd podchaos.chaos-mesh.org
    ```

2. Install Chaos Mesh:

    ```bash
    # for helm v2
    helm install helm/chaos-mesh --name=chaos-mesh --namespace=chaos-testing --set chaosDaemon.runtime=containerd --set chaosDaemon.socketPath=/run/containerd/containerd.sock
    # for helm v3
    helm install chaos-mesh helm/chaos-mesh --namespace=chaos-testing --set chaosDaemon.runtime=containerd --set chaosDaemon.socketPath=/run/containerd/containerd.sock
    kubectl get pods --namespace chaos-testing -l app.kubernetes.io/instance=chaos-mesh
    ```

    The arguments `--set chaosDaemon.runtime=containerd --set chaosDaemon.socketPath=/run/containerd/containerd.sock` is used to to support network chaos on kind.

3. Create `chaos.yaml` in any location with the lines below:

    ```yaml
    apiVersion: chaos-mesh.org/v1alpha1
    kind: HelloWorldChaos
    metadata:
      name: hello-world
      namespace: chaos-testing
    spec: {}
    ```

4. Apply the chaos:

    ```bash
    kubectl apply -f /path/to/chaos.yaml
    kubectl get HelloWorldChaos -n chaos-testing
    ```

    Now you should be able to check the `Hello World!` result in the log:

    ```bash
    kubectl logs chaos-controller-manager-{pod-post-fix} -n chaos-testing
    ```

    > **Note:**
    >
    > `{pod-post-fix}` is a random string generated by Kubernetes, you can check it by executing `kubectl get po -n chaos-testing`.

## Next steps

Congratulations! You have just added a chaos type for Chaos Mesh successfully. Let us know if you run into any issues during the process. If you feel like doing other types of contributions, refer to Add facilities to chaos daemon (WIP).
