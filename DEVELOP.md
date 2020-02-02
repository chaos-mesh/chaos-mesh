# Chaos Mesh Development Guide
Chaos Mesh currently have many types of chaos such as IOChaos, NetworkChaos, PodChaos. This guide aims to introduce the way to add new chaos type for Chaos Mesh, which includes setting up the development environment and adding a hello-world chaos type.

## Setting Up The Development Environment
### Requirements
You should have these components installed in your system:
- golang (>= v1.13), if your golang version < v1.13, you can use golang version manager such as [gvm](https://github.com/moovweb/gvm) to switch to a newer one.
- [yarn](https://yarnpkg.com/lang/en/) and [nodejs](https://nodejs.org/en/): for chaos-dashboard
- docker
- gcc
- [kind](https://github.com/kubernetes-sigs/kind)

### Step By Step
Firstly, clone the repo
```
git clone https://github.com/pingcap/chaos-mesh.git
cd chaos-mesh
```
Then, install [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) and [kustomize](https://github.com/kubernetes-sigs/kustomize)
```
make install-kubebuilder
make install-kustomize
```
After this, you should add kubebuilder and kustomize to your PATH environment:
```
export PATH=${PATH}:/usr/local/kubebuilder/bin
```

start docker service, if you are using centos, you can use the following command to start
```
service docker start
```
Then we can test the toolchain
```
make
```
It should work. But it's not enough yet, we need to set up a local Kubernetes cluster, which needs [kind](https://github.com/kubernetes-sigs/kind)
```
curl -Lo ./kind "https://github.com/kubernetes-sigs/kind/releases/download/v0.7.0/kind-$(uname)-amd64"
chmod +x ./kind
mv kind /usr/local/bin/
```
In addition to kind, you also need helm
```
curl -Lo helm.tar.gz https://get.helm.sh/helm-v2.15.1-linux-amd64.tar.gz
tar xvf helm.tar.gz
mv linux-amd64/helm /usr/local/bin/
```
Then we can set up the k8s cluster with the script
```
hack/kind-cluster-build.sh
```
Now we have our environment done!

## Develop A HelloWorldChaos
After preparing the development environment, let's develop a new type of chaos, HelloWorldChaos, which only prints a "hello world" message to log.

As we know, the chaos is managed by the controller manager, so we should do something with the controller manager to add our HelloWorldChaos. 

Check out the [main.go](https://github.com/pingcap/chaos-mesh/blob/master/cmd/controller-manager/main.go#L104) of controller manager we can see there are already some types of chaos:
```golang
	if err = (&controllers.PodChaosReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("PodChaos"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "PodChaos")
		os.Exit(1)
	}

	if err = (&controllers.NetworkChaosReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("NetworkChaos"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "NetworkChaos")
		os.Exit(1)
	}

	if err = (&controllers.IoChaosReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("IoChaos"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "IoChaos")
		os.Exit(1)
	}
```
There are PodChaos, NetworkChaos and IoChaos. Let's add HelloWorldChaos:
```golang
	if err = (&controllers.HelloWorldChaosReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("HelloWorldChaos"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "HelloWorldChaos")
		os.Exit(1)
	}
```
Of course this won't work since we didn't implement HelloWorldChaosReconciler yet, we should implement it in [controllers](https://github.com/pingcap/chaos-mesh/tree/master/controllers).

Add a new file helloworldchaos_controller.go:
```golang
package controllers

import (
	"github.com/go-logr/logr"

	chaosmeshv1alpha1 "github.com/pingcap/chaos-mesh/api/v1alpha1"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// HelloWorldChaosReconciler reconciles a HelloWorldChaos object
type HelloWorldChaosReconciler struct {
	client.Client
	Log logr.Logger
}

// +kubebuilder:rbac:groups=pingcap.com,resources=helloworldchaos,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=pingcap.com,resources=helloworldchaos/status,verbs=get;update;patch

func (r *HelloWorldChaosReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	logger := r.Log.WithValues("reconciler", "helloworldchaos")

        // This is what we want to do
	logger.Info("Hello World!")

	return ctrl.Result{}, nil
}

func (r *HelloWorldChaosReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&chaosmeshv1alpha1.HelloWorldChaos{}).
		Complete(r)
}
```
We implement a reconciler in this file, it has only two methods:
- `Reconcile`: the main logic of `HelloWorldChaos`, it prints a log `Hello World!` and returns nothing.
- `SetupWithManager`: as you see in [main.go](https://github.com/pingcap/chaos-mesh/blob/master/cmd/controller-manager/main.go#L104), this method is used to export `HelloWorldChaos` object, which represents the yaml schema content the user applies.

The comment `// +kubebuilder:rbac:groups=pingcap.com...` is an authority control mechanism that decides which account can access this reconciler. To make it be accessible by dashboard and chaos-controller-manager we should modify [collector-rbac.yaml](https://github.com/pingcap/chaos-mesh/blob/master/helm/chaos-mesh/templates/collector-rbac.yaml) and [controller-manager-rbac.yaml](https://github.com/pingcap/chaos-mesh/blob/master/helm/chaos-mesh/templates/controller-manager-rbac.yaml), adding helloworldchaos to resources of all "pingcap.com" apiGroup:
```yaml
  - apiGroups: ["pingcap.com"]
    resources:
      - podchaos
      - networkchaos
      - iochaos
      - helloworldchaos    # Add this line in all pingcap.com group
    verbs: ["*"]
```
HelloWorldChaos is a CRD in k8s, so we should register it, to do this, we can modify [kustomization.yaml](https://github.com/pingcap/chaos-mesh/blob/master/config/crd/kustomization.yaml), adding one line in resources section:
```yaml
resources:
- bases/pingcap.com_podchaos.yaml
- bases/pingcap.com_networkchaos.yaml
- bases/pingcap.com_iochaos.yaml
- bases/pingcap.com_helloworldchaos.yaml  # this is the new line
```

The last step is to implement the schema type in [api directory](https://github.com/pingcap/chaos-mesh/tree/master/api/v1alpha1). Add helloworldchaos_types.go:
```golang
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true

// HelloWorldChaos is the Schema for the helloworldchaos API
type HelloWorldChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
}

// +kubebuilder:object:root=true

// HelloWorldChaosList contains a list of HelloWorldChaos
type HelloWorldChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HelloWorldChaos `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HelloWorldChaos{}, &HelloWorldChaosList{})
}
```
The `HelloWorldChaos` type represents the yaml content:
```yaml
apiVersion: pingcap.com/v1alpha1
kind: HelloWorldChaos
metadata:
  name: <name-of-this-resource>
  namespace: <ns-of-this-resource>
```

Having all these done, we can make images now:
```
make
make docker-push
```
Note that the default `DOCKER_REGISTRY` is `localhost:5000`, which is set up by `hack/kind-cluster-build.sh`, you can overwrite it to any registry you have access permission, for development, this is enough.

Before we install or upgrade chaos-mesh, we should modify [values.yaml](https://github.com/pingcap/chaos-mesh/blob/master/helm/chaos-mesh/values.yaml) of helm template, replacing the image to what we have pushed. For example, the template using `pingcap/chaos-mesh:latest` as the target image, but what we have developed is `localhost:5000/pingcap/chaos-mesh:latest`, so we should correct this:
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
Now we can create the related custom resource type for chaos-mesh:
```
kubectl apply -f manifests/
kubectl get crd podchaos.pingcap.com
```
Install Chaos Mesh:
```
helm install helm/chaos-mesh --name=chaos-mesh --namespace=chaos-testing --set chaosDaemon.runtime=containerd --set chaosDaemon.socketPath=/run/containerd/containerd.sock
kubectl get pods --namespace chaos-testing -l app.kubernetes.io/instance=chaos-mesh
```
The arguments `--set chaosDaemon.runtime=containerd --set chaosDaemon.socketPath=/run/containerd/containerd.sock` is used to to support network chaos on kind.

Create chaos.yaml with content:
```yaml
apiVersion: pingcap.com/v1alpha1
kind: HelloWorldChaos
metadata:
  name: hello-world
  namespace: chaos-testing
```
And apply it:
```
kubectl apply -f chaos.yaml
kubectl get HelloWorldChaos -n chaos-testing
```
And try to check out the `Hello World!` result:
```
kubectl logs chaos-controller-manager-{pod-post-fix} -n chaos-testing
```
The `pod-post-fix` is the random string generated by k8s, my pod name is `chaos-controller-manager-7fcc54c658-gnkjk`, you can check out yours by `kubectl get pods --namespace chaos-testing|grep chaos-controller-manager`.

With the command `kubectl logs ...`, you should see the `Hello World!`.
