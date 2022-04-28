# Chaos Mesh

[Chaos Mesh](https://github.com/chaos-mesh/chaos-mesh) is a cloud-native Chaos Engineering platform that orchestrates chaos on Kubernetes environments.

## Introduction

This chart bootstraps a [Chaos Mesh](https://github.com/chaos-mesh/chaos-mesh) deployment on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.

## Deploy

Before deploying Chaos Mesh, make sure you have installed the [Prerequisites](https://chaos-mesh.org/docs/production-installation-using-helm#prerequisites). And then follow the [install-by-helm](https://chaos-mesh.org/docs/production-installation-using-helm#install-chaos-mesh-using-helm) doc step by step.

## Configuration

The following tables list the configurable parameters of the Chaos Mesh chart and their default values.

|                 Parameter                  |                                                     Description                                                      |                         Default                         |
|--------------------------------------------|----------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------|
| `nameOverride` |  | `` |
| `fullnameOverride` |  | `` |
| `customLabels` | Customized labels that will be tagged on all the resources of Chaos Mesh | `{}` |
| `clusterScoped` | Whether chaos-mesh should manage kubernetes cluster wide chaos.Also see rbac.create and controllerManager.serviceAccount | `true` |
| `rbac.create` | Creating rbac API Objects. Also see clusterScoped and controllerManager.serviceAccount | `true`                                                |
| `timezone` | The timezone where controller-manager, chaos-daemon and dashboard uses. For example: `UTC`, `Asia/Shanghai` | `UTC` |
| `enableProfiling` | A flag to enable pprof in controller-manager and chaos-daemon  | `true` |
| `enableCtrlServer` | A flag to enable ctrlserver which provides service to chaosctl in controller-manager. | `true` |
| `images.registry` | The global container registry for the images, you could replace it with your self-hosted container registry. | `ghcr.io` |
| `images.tag` | The global image tag (for example, semiVer with prefix v, or latest). | `latest` |
| `imagePullSecrets` | Global Docker registry secret names as an array  | [] (does not add image pull secrets to deployed pods) |
| `controllerManager.hostNetwork` | Running chaos-controller-manager on host network | `false` |
| `controllerManager.allowHostNetworkTesting`   | Allow testing on `hostNetwork` pods | `false` |
| `controllerManager.serviceAccount` | The serviceAccount for chaos-controller-manager | `chaos-controller-manager` |
| `controllerManager.priorityClassName` | Custom priorityClassName for using pod priorities | `` |
| `controllerManager.replicaCount` | Replicas for chaos-controller-manager | `3` |
| `controllerManager.image.registry` | Override global registry, empty value means using the global images.registry | `` |
| `controllerManager.image.repository` | Repository part for image of chaos-controller-manager | `chaos-mesh/chaos-mesh` |
| `controllerManager.image.tag` | Override global tag, empty value means using the global images.tag | `` |
| `controllerManager.imagePullPolicy` | Image pull policy | `Always` |
| `controllerManager.enableFilterNamespace` | If enabled, only pods in the namespace annotated with `"chaos-mesh.org/inject": "enabled"` could be injected | false |
| `controllerManager.service.type` | Kubernetes Service type for service chaos-controller-manager | `ClusterIP` |
| `controllerManager.resources` | CPU/Memory resource requests/limits for chaos-controller-manager pod | `{requests: { cpu: "25m", memory: "256Mi" }, limits:{}}`   |
| `controllerManager.nodeSelector` |  Node labels for chaos-controller-manager pod assignment | `{}` |
| `controllerManager.tolerations` |  Toleration labels for chaos-controller-manager pod assignment | `[]` |
| `controllerManager.affinity` |  Map of chaos-controller-manager node/pod affinities | `{}` |
| `controllerManager.podAnnotations` |  Pod annotations of chaos-controller-manager | `{}`|
| `controllerManager.enabledControllers`| A list of controllers to enable. "*" enables all controllers by default. | `["*"]` |
| `controllerManager.enabledWebhooks`| A list of webhooks to enable. "*" enables all webhooks by default. | `["*"]` |
| `controllerManager.podChaos.podFailure.pauseImage` | Custom Pause Container Image for Pod Failure Chaos | `gcr.io/google-containers/pause:latest` |
| `controllerManager.leaderElection.enabled` | Enable leader election for controller manager. | `true` |
| `controllerManager.leaderElection.leaseDuration` | The duration that non-leader candidates will wait to force acquire leadership. This is measured against time of last observed ack. | `15s` |
| `controllerManager.leaderElection.renewDeadline` | The duration that the acting control-plane will retry refreshing leadership before giving up. | `10s` |
| `controllerManager.leaderElection.retryPeriod` | The duration the LeaderElector clients should wait between tries of actions. | `2s` |
| `controllerManager.chaosdSecurityMode` |  Enabled for mTLS connection between chaos-controller-manager and chaosd | `true` |
| `controllerManager.image.registry` | Override global registry, empty value means using the global images.registry | `` |
| `controllerManager.image.repository` | Repository part for image of chaos-daemon | `chaos-mesh/chaos-daemon` |
| `controllerManager.image.tag` | Override global tag, empty value means using the global images.tag | `` |
| `chaosDaemon.imagePullPolicy` | Image pull policy | `Always` |
| `chaosDaemon.grpcPort` | The port which grpc server listens on | `31767` |
| `chaosDaemon.httpPort` | The port which http server listens on | `31766` |
| `chaosDaemon.env` | Extra chaosDaemon envs | `{}` |
| `chaosDaemon.hostNetwork` | Running chaosDaemon on host network | `false` |
| `chaosDaemon.mtls.enabled` | Enable mtls on the grpc connection between chaos-controller-manager and chaos-daemon | `true` |
| `chaosDaemon.privileged` | Run chaos-daemon container in privileged mode. If it is set to false, chaos-daemon will be run in some specified capabilities. capabilities: SYS_PTRACE, NET_ADMIN, MKNOD, SYS_CHROOT, SYS_ADMIN, KILL, IPC_LOCK | `true` |
| `chaosDaemon.priorityClassName` | Custom priorityClassName for using pod priorities | `` |
| `chaosDaemon.podAnnotations` | Pod annotations of chaos-daemon | `{}` |
| `chaosDaemon.serviceAccount`| ServiceAccount name for chaos-daemon | `chaos-daemon` |
| `chaosDaemon.podSecurityPolicy` | Specify PodSecurityPolicy(psp) on chaos-daemon pods | `false`|
| `chaosDaemon.runtime` | Runtime specifies which container runtime to use. Currently we only supports docker, containerd and CRI-O. | `docker` |
| `chaosDaemon.socketPath` | Specifiesthe path of container runtime socket on the host. | `/var/run/docker.sock` |
| `chaosDaemon.resources` | CPU/Memory resource requests/limits for chaosDaemon container | `{}`  |
| `chaosDaemon.nodeSelector` |  Node labels for chaos-daemon pod assignment | `{}` |
| `chaosDaemon.tolerations` |  Toleration labels for chaos-daemon pod assignment | `[]` |
| `chaosDaemon.affinity` |  Map of chaos-daemon node/pod affinities | `{}` |
| `dashboard.create` | Enable chaos-dashboard | `false` |
| `dashboard.rootUrl` | Specify the base url for openid/oauth2 (like GCP Auth Integration) callback URL. | `http://localhost:2333` |
| `dashboard.hostNetwork` | Running chaos-dashboard on host network | `false` |
| `dashboard.replicaCount` | Replicas of chaos-dashboard | `1` |
| `dashboard.priorityClassName` | Custom priorityClassName for using pod priorities | `` |
| `dashboard.serviceAccount` | The serviceAccount for chaos-dashboard  | `chaos-dashboard` |
| `dashboard.image.registry` | Override global registry, empty value means using the global images.registry | `` |
| `dashboard.image.repository` | Repository part for image of chaos-dashboard | `chaos-mesh/chaos-dashboard` |
| `dashboard.image.tag` | Override global tag, empty value means using the global images.tag | `` |
| `dashboard.imagePullPolicy` | Image pull policy | `Always` |
| `dashboard.securityMode` | Require user to provide credentials on Chaos Dashboard, instead of using chaos-dashboard service account | `true` |
| `dashboard.gcpSecurityMode` | Enable GCP Authentication Integration, see: <https://chaos-mesh.org/docs/gcp-authentication/> for more details | `false` |
| `dashboard.gcpClientId` | GCP app's client ID with GCP Authentication Integration | `` |
| `dashboard.gcpClientSecret` | GCP app's client secret with GCP Authentication Integration | `` |
| `dashboard.nodeSelector` | Node labels for chaos-dashboard  pod assignment | `{}` |
| `dashboard.tolerations` | Toleration labels for chaos-dashboard pod assignment | `[]` |
| `dashboard.affinity` | Map of chaos-dashboard node/pod affinities | `{}` |
| `dashboard.podAnnotations` | Deployment chaos-dashboard annotations | `{}` |
| `dashboard.service.annotations` | Service annotations for the dashboard | `{}` |
| `dashboard.service.type`              | Service type of the service created for exposing the dashboard                             | `NodePort`     |
| `dashboard.service.clusterIP`         | Set the `clusterIP` of the dashboard service if the type is `ClusterIP` | `nil`           |
| `dashboard.service.nodePort`          | Set the `nodePort` of the dashboard service if the type is `NodePort`  | `nil`           |
| `dashboard.resources` | CPU/Memory resource requests/limits for chaos-dashboard pod  | `requests: { cpu: "25m", memory: "256Mi" }, limits:{}`  |
| `dashboard.persistentVolume.enable` | Enable storage volume for chaos-dashboard. If you are using SQLite as your DB for Chaos Dashboard, it is recommended to enable persistence| `false` |
| `dashboard.persistentVolume.existingClaim` | Use the existing PVC for persisting chaos event| `` |
| `dashboard.persistentVolume.size` | Chaos Dashboard data Persistent Volume size | `8Gi` |
| `dashboard.persistentVolume.storageClassName` | Chaos Dashboard data Persistent Volume Storage Class | `standard` |
| `dashboard.persistentVolume.mountPath` | Chaos Dashboard data Persistent Volume mount root path | `/data` |
| `dashboard.persistentVolume.subPath` | Subdirectory of  Chaos Dashboard data Persistent Volume to mount | `` |
| `dashboard.env` | The keys within the `env` map are mounted as environment variables on the Chaos Dashboard pod | `` |
| `dashboard.env.LISTEN_HOST` | The address which chaos-dashboard would listen on. | `0.0.0.0` |
| `dashboard.env.LISTEN_PORT` | The port which chaos-dashboard would listen on. | `2333` |
| `dashboard.env.METRIC_HOST` | The address which metrics endpoints would listen on. | `0.0.0.0` |
| `dashboard.env.METRIC_PORT` | The address which metrics endpoints would listen on. | `2334` |
| `dashboard.env.DATABASE_DRIVER`| The db drive used for Chaos Dashboard, support db: sqlite3, mysql| `sqlite3` |
| `dashboard.env.DATABASE_DATASOURCE`| The db dsn used for Chaos Dashboard | `/data/core.sqlite` |
| `dashboard.env.CLEAN_SYNC_PERIOD`| Set the sync period to clean up archived data | `12h` |
| `dashboard.env.TTL_EVENT`| Set TTL of archived event data | `168h` |
| `dashboard.env.TTL_EXPERIMENT`| Set TTL of archived experiment data | `336h` |
| `dashboard.env.TTL_SCHEDULE`| Set TTL of archived schedule data | `336h` |
| `dashboard.env.TTL_WORKFLOW`| Set TTL of archived workflow data | `336h` |
| `dashboard.ingress.enabled`                   | Enable the use of the ingress controller to access the dashboard                         | `false`             |
| `dashboard.ingress.certManager`               | Enable Cert-Manager for ingress                                                      | `false`             |
| `dashboard.ingress.annotations`               | Annotations for the dashboard Ingress                                                   | `{}`                |
| `dashboard.ingress.hosts[0].name`             | Hostname to your dashboard installation                                                 | `dashboard.local`     |
| `dashboard.ingress.hosts[0].paths`            | Path within the url structure                                                         | `["/"]`             |
| `dashboard.ingress.hosts[0].tls`              | Utilize TLS backend in ingress                                                        | `false`             |
| `dashboard.ingress.hosts[0].tlsHosts`         | Array of TLS hosts for ingress record (defaults to `ingress.hosts[0].name` if `nil`)  | `nil`               |
| `dashboard.ingress.hosts[0].tlsSecret`        | TLS Secret (certificates)                                                             | `dashboard.local-tls` |
| `dnsServer.create` | Enable DNS Server which required by DNSChaos | `false` |
| `dnsServer.serviceAccount` | Name of serviceaccount for chaos-dns-server. | `chaos-dns-server` |
| `dnsServer.image` | Image of DNS Server | `pingcap/coredns:v0.2.1` |
| `dnsServer.imagePullPolicy` | Image pull policy | `IfNotPresent` |
| `dnsServer.priorityClassName` | Customized priorityClassName for chaos-dns-server | `` |
| `dnsServer.nodeSelector` | Node labels for chaos-dns-server pod assignment | `` |
| `dnsServer.tolerations` | Toleration labels for chaos-dns-server pod assignment | `[]` |
| `dnsServer.podAnnotations` | Pod annotations of chaos-dns-server | `` |
| `dnsServer.name` | The service name of chaos-dns-server | `chaos-mesh-dns-server` |
| `dnsServer.grpcPort` | Grpc port for chaos-dns-server | `9288` |
| `dnsServer.resources` | CPU/Memory resource requests/limits for chaos-dns-server pod |  `requests: { cpu: "100m", memory: "70Mi" }, limits:{}` |
| `dnsServer.env.LISTEN_HOST` | The address of chaos-dns-server listen on | `0.0.0.0` |
| `dnsServer.env.LISTEN_PORT` | The port of chaos-dns-server listen on | `53` |
| `prometheus.create` | Enable prometheus | `false` |
| `prometheus.serviceAccount` | The serviceAccount for prometheus | `prometheus` |
| `prometheus.image` | Docker image for prometheus | `prom/prometheus:v2.15.2` |
| `prometheus.imagePullPolicy` | Image pull policy | `IfNotPresent` |
| `prometheus.priorityClassName` | Custom priorityClassName for using pod priorities | `` |
| `prometheus.nodeSelector` | Node labels for prometheus pod assignment | `{}` |
| `prometheus.tolerations` | Toleration labels for prometheus pod assignment | `[]` |
| `prometheus.affinity` | Map of prometheus node/pod affinities | `{}` |
| `prometheus.podAnnotations` | Deployment prometheus annotations | `{}` |
| `prometheus.resources` | CPU/Memory resource requests/limits for prometheus pod |  `requests: { cpu: "250m", memory: "512Mi" }, limits:{ cpu: "500m", memory: "1024Mi" }`  |
| `prometheus.service.type` | Kubernetes Service type | `ClusterIP` |
| `prometheus.volume.storage` | Storage size of PVC | `2Gi` |
| `prometheus.volume.storageClassName` | Storage class of PVC | `standard` |
| `webhook.certManager.enabled` | Setup the webhook using cert-manager | `false` |
| `webhook.timeoutSeconds` | Timeout for admission webhooks in seconds | `5` |
| `webhook.FailurePolicy` | Defines how unrecognized errors and timeout errors from the admission webhook are handled | `Fail` |
| `webhook.CRDS` | Define a list of chaos types that implement admission webhook | `[podchaos,iochaos,timechaos,networkchaos,kernelchaos,stresschaos,awschaos,azurechaos,gcpchaos,dnschaos,jvmchaos,schedule,workflow,httpchaos,bnlockchaos,physicalmachinechaos,phsicalmachine,statuscheck]` |
| `bpfki.create` | Enable chaos-kernel | `false` |
| `bpfki.image.registry` | Override global registry, empty value means using the global images.registry | `` |
| `bpfki.image.repository` | Repository part for image of chaos-kernel | `chaos-mesh/chaos-kernel` |
| `bpfki.image.tag` | Override global tag, empty value means using the global images.tag | `` |
| `bpfki.imagePullPolicy` | Image pull policy | `IfNotPresent` |
| `bpfki.grpcPort` | The port which grpc server listens on | `50051` |
| `bpfki.resources` | CPU/Memory resource requests/limits for chaos-kernel container | {}}` |
| `chaosDlv.enable` | Create sidecar remote debugging container | `false` |
| `chaosDlv.image.registry` | Override global registry, empty value means using the global images.registry | `false` |
| `chaosDlv.repository` | Repository part for image of chaos-dlv | `chaos-mesh/chaos-dlv` |
| `chaosDlv.tag` | Override global tag, empty value means using the global images.tag | `false` |
| `chaosDlv.imagePullPolicy` | Image pull policy | `IfNotPresent` |

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

[Cert-manager](https://github.com/jetstack/cert-manager) may be the default in the K8s world for certificate management now. If you want to install Cert-manager using the [Helm](https://helm.sh) package manager, please refer to the [official documents](https://github.com/jetstack/cert-manager/tree/master/deploy/charts/cert-manager).

Example for deploy Cert-manager

```bash
helm repo add jetstack https://charts.jetstack.io
helm repo update

# if Kubernetes > 1.18/Helm 3.2
helm install cert-manager jetstack/cert-manager --namespace cert-manager --create-namespace --version v1.6.1 --set installCRDs=true

# else
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.6.1/cert-manager.crds.yaml
helm install cert-manager jetstack/cert-manager --namespace cert-manager --create-namespace --version v1.6.1
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

> <https://github.com/jetstack/cert-manager/issues/296>
>
> <https://github.com/jetstack/cert-manager/pull/819>

You can install your Cert-manager looks like this.

```bash
helm install cert-manager jetstack/cert-manager --namespace cert-manager --version v0.13.1 --set extraArgs={"--enable-certificate-owner-ref"="true"}
```
