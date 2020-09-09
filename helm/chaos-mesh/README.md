# Chaos Mesh

[Chaos Mesh](https://github.com/chaos-mesh/chaos-mesh) is a cloud-native Chaos Engineering platform that orchestrates chaos on Kubernetes environments.

## Introduction

This chart bootstraps a [Chaos Mesh](https://github.com/chaos-mesh/chaos-mesh) deployment on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.

## Deploy

Before deploying Chaos Mesh, make sure you have installed the [Prerequisites](../../website/docs/installation/installation.md#prerequisites). And then follow the [install-by-helm](../../website/docs/installation/installation.md#install-by-helm) doc step by step.

## Configuration

The following tables list the configurable parameters of the Chaos Mesh chart and their default values.

|                 Parameter                  |                                                     Description                                                      |                         Default                         |
|--------------------------------------------|----------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------|
| `clusterScoped`                            | whether chaos-mesh should manage kubernetes cluster wide chaos.Also see rbac.create and controllerManager.serviceAccount | `true` |
| `rbac.create` |  | `true`                                                |
| `timezone` | The timezone where controller-manager, chaos-daemon and dashboard uses. For example: `UTC`, `Asia/Shanghai` | `UTC` |
| `enableProfiling` | A flag to enable pprof in controller-manager and chaos-daemon  | `true` |
| `controllerManager.hostNetwork` | running chaos-controller-manager on host network | `false` |
| `controllerManager.serviceAccount` | The serviceAccount for chaos-controller-manager | `chaos-controller-manager` |
| `controllerManager.replicaCount` | Replicas for chaos-controller-manager | `1` |
| `controllerManager.image` | docker image for chaos-controller-manager  | `pingcap/chaos-mesh:latest` |
| `controllerManager.imagePullPolicy` | Image pull policy | `Always` |
| `controllerManager.nameOverride` |  |  |
| `controllerManager.fullnameOverride` |  |  |
| `controllerManager.service.type` | Kubernetes Service type | `ClusterIP` |
| `controllerManager.resources` | CPU/Memory resource requests/limits for chaos-controller-manager pod | `requests: { cpu: "250m", memory: "512Mi" }, limits:{ cpu: "500m", memory: "1024Mi" }`   |
| `controllerManager.nodeSelector` |  Node labels for chaos-controller-manager pod assignment | `{}` |
| `controllerManager.tolerations` |  Toleration labels for chaos-controller-manager pod assignment | `[]` |
| `controllerManager.affinity` |  Map of chaos-controller-manager node/pod affinities | `{}` |
| `controllerManager.podAnnotations` |  Pod annotations of chaos-controller-manager | `{}`|
| `controllerManager.allowedNamespaces` |  A regular expression, and matching namespace will allow the chaos task to be performed | ``|
| `controllerManager.ignoredNamespaces` |  A regular expression, and the chaos task will be ignored by a matching namespace. Configuring `allowedNamespaces` at the same time will ignore this configuration. | ``|
| `chaosDaemon.image` | docker image for chaos-daemon | `pingcap/chaos-mesh:latest` |
| `chaosDaemon.imagePullPolicy` | image pull policy | `Always` |
| `chaosDaemon.grpcPort` | The port which grpc server listens on | `31767` |
| `chaosDaemon.httpPort` | The port which http server listens on | `31766` |
| `chaosDaemon.env` | chaosDaemon envs | `{}` |
| `chaosDaemon.hostNetwork` | running chaosDaemon on host network | `false` |
| `chaosDaemon.podAnnotations` | Pod annotations of chaos-daemon | `{}` |
| `chaosDaemon.runtime` | Runtime specifies which container runtime to use. Currently we only supports docker and containerd. | `docker` |
| `chaosDaemon.socketPath` | Specifies the container runtime socket | `/var/run/docker.sock` |
| `chaosDaemon.tolerations` | Toleration labels for chaos-daemon pod assignment | `[]` |
| `chaosDaemon.resources` | CPU/Memory resource requests/limits for chaosDaemon container | `requests: { cpu: "250m", memory: "512Mi" }, limits:{ cpu: "500m", memory: "1024Mi" }`  |
| `bpfki.create` | Enable chaos-kernel | `false` |
| `bpfki.image` | Docker image for chaos-kernel | `pingcap/chaos-kernel:latest` |
| `bpfki.imagePullPolicy` | Image pull policy | `Always` |
| `bpfki.grpcPort` | The port which grpc server listens on | `50051` |
| `bpfki.resources` | CPU/Memory resource requests/limits for chaos-kernel container | `requests: { cpu: "250m", memory: "512Mi" }, limits:{ cpu: "500m", memory: "1024Mi" }`  |
| `dashboard.create` | Enable chaos-dashboard | `false` |
| `dashboard.serviceAccount` | The serviceAccount for chaos-dashboard  | `chaos-dashboard` |
| `dashboard.image` | Docker image for chaos-dashboard | `pingcap/chaos-dashboard:latest` |
| `dashboard.imagePullPolicy` | Image pull policy | `Always` |
| `dashboard.nodeSelector` | Node labels for chaos-dashboard  pod assignment | `{}` |
| `dashboard.tolerations` | Toleration labels for chaos-dashboard pod assignment | `[]` |
| `dashboard.affinity` | Map of chaos-dashboard node/pod affinities | `{}` |
| `dashboard.podAnnotations` | Deployment chaos-dashboard annotations | `{}` |
| `dashboard.resources` | CPU/Memory resource requests/limits for chaos-dashboard pod  | `requests: { cpu: "250m", memory: "512Mi" }, limits:{ cpu: "500m", memory: "1024Mi" }`  |
| `dashboard.persistentVolume.enable` | Enable storage volume for chaos-dashboard. If you are using SQLite as your DB for Chaos Dashboard, it is recommended to enable persistence| `false` |
| `dashboard.persistentVolume.existingClaim` | Use the existing PVC for persisting chaos event| `` |
| `dashboard.persistentVolume.size` | Chaos Dashboard data Persistent Volume size | `8Gi` |
| `dashboard.persistentVolume.storageClassName` | Chaos Dashboard data Persistent Volume Storage Class | `standard` |
| `dashboard.persistentVolume.mountPath` | Chaos Dashboard data Persistent Volume mount root path | `/data` |
| `dashboard.persistentVolume.subPath` | Subdirectory of  Chaos Dashboard data Persistent Volume to mount | `` |
| `dashboard.env` | The keys within the `env` map are mounted as environment variables on the Chaos Dashboard pod | `` |
| `dashboard.env.LISTEN_HOST` | | `0.0.0.0` |
| `dashboard.env.LISTEN_PORT` | | `2333` |
| `dashboard.env.DATABASE_DRIVER`| The db drive used for Chaos Dashboard, support db: sqlite3, mysql| `sqlite3` |
| `dashboard.env.DATABASE_DATASOURCE`| The db dsn used for Chaos Dashboard | `/data/core.sqlite` |
| `dashboard.ingress.enabled`                   | Enable the use of the ingress controller to access the dashboard                         | `false`             |
| `dashboard.ingress.certManager`               | Enable Cert-Manager for ingress                                                      | `false`             |
| `dashboard.ingress.annotations`               | Annotations for the dashboard Ingress                                                   | `{}`                |
| `dashboard.ingress.hosts[0].name`             | Hostname to your dashboard installation                                                 | `dashboard.local`     |
| `dashboard.ingress.hosts[0].paths`            | Path within the url structure                                                         | `["/"]`             |
| `dashboard.ingress.hosts[0].tls`              | Utilize TLS backend in ingress                                                        | `false`             |
| `dashboard.ingress.hosts[0].tlsHosts`         | Array of TLS hosts for ingress record (defaults to `ingress.hosts[0].name` if `nil`)  | `nil`               |
| `dashboard.ingress.hosts[0].tlsSecret`        | TLS Secret (certificates)                                                             | `dashboard.local-tls` |
| `prometheus.create` | Enable prometheus | `false` |
| `prometheus.serviceAccount` | The serviceAccount for prometheus | `prometheus` |
| `prometheus.image` | Docker image for prometheus | `prom/prometheus:v2.15.2` |
| `prometheus.imagePullPolicy` | Image pull policy | `IfNotPresent` |
| `prometheus.nodeSelector` | Node labels for prometheus pod assignment | `{}` |
| `prometheus.tolerations` | Toleration labels for prometheus pod assignment | `[]` |
| `prometheus.affinity` | Map of prometheus node/pod affinities | `{}` |
| `prometheus.podAnnotations` | Deployment prometheus annotations | `{}` |
| `prometheus.resources` | CPU/Memory resource requests/limits for prometheus pod |  `requests: { cpu: "250m", memory: "512Mi" }, limits:{ cpu: "500m", memory: "1024Mi" }`  |
| `prometheus.service.type` | Kubernetes Service type | `ClusterIP` |
| `prometheus.volume.storage` | | `2Gi` |
| `prometheus.volume.storageClassName` | | `standard` |
| `webhook.certManager.enabled` | Setup the webhook using cert-manager | `false` |
| `webhook.FailurePolicy` | Defines how unrecognized errors and timeout errors from the admission webhook are handled | `Ignore` |
| `webhook.CRDS` | Define a list of chaos types that implement admission webhook | `[podchaos,iochaos,timechaos,networkchaos,kernelchaos]` |

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example,
```console
# helm 2.X
helm install helm/chaos-mesh --name=chaos-mesh --namespace=chaos-testing --set dashboard.create=true
# helm 3.X
helm install chaos-mesh helm/chaos-mesh --namespace=chaos-testing --set dashboard.create=true
```

The above command enable the Chaos Dashboard.

Alternatively, a YAML file that specifies the values for the parameters can be provided while installing the chart. For example,

```console
# helm 2.X
helm install helm/chaos-mesh --name=chaos-mesh --namespace=chaos-testing -f values.yaml
# helm 3.X
helm install chaos-mesh helm/chaos-mesh --namespace=chaos-testing -f values.yaml
```

> **Tip**: You can use the default [values.yaml](values.yaml)

## Configuration and installation details

### Using cert-manager for certificate management

[Cert-manager](https://github.com/jetstack/cert-manager) may be the default in the K8s world for certificate management now.If you want to install Cert-manager using the [Helm](https://helm.sh) package manager, please refer to the [official documents](https://github.com/jetstack/cert-manager/tree/master/deploy/charts/cert-manager).

Example for deploy Cert-manager

```bash
kubectl create namespace cert-manager
kubectl apply --validate=false -f https://raw.githubusercontent.com/jetstack/cert-manager/v0.13.1/deploy/manifests/00-crds.yaml
helm repo add jetstack https://charts.jetstack.io
helm repo update
helm install cert-manager jetstack/cert-manager --namespace cert-manager --version v0.13.1
```

In case you want to using Cert-manager for certificate management, you can use the `webhook.certManager.enabled` property.

```yaml
webhook:
  certManager:
    enabled: true
```

The webhook's cert and the [MutatingAdmissionWebhook](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#mutatingadmissionwebhook)'s `caBundle` property will be managed by the [Certificate](https://cert-manager.io/docs/usage/certificate/) of Cert-manager.

In case your Cert-manager's option `enable-certificate-owner-ref` is true, it means that deleting a certificate resource will also delete its secret.

The Cert-manager's option `enable-certificate-owner-ref` refer to the following:
> https://github.com/jetstack/cert-manager/issues/296
>
> https://github.com/jetstack/cert-manager/pull/819

You can install your Cert-manager looks like this.

```bash
helm install cert-manager jetstack/cert-manager --namespace cert-manager --version v0.13.1 --set extraArgs={"--enable-certificate-owner-ref"="true"}
```
