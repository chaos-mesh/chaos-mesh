# Set up the development environment
### Prerequisites
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


Install [docker](https://docs.docker.com/install/) and start docker service, if you are using centos, you can use the following command to start
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
