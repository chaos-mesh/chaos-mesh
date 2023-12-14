#!/usr/bin/env bash
# Copyright 2021 Chaos Mesh Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# This is a script to quickly install chaos-mesh.
# This script will check if docker and kubernetes are installed. If local mode is set and kubernetes is not installed,
# it will use kind or minikube to install the kubernetes cluster according to the configuration.
# Finally, when all dependencies are installed, chaos-mesh will be installed.

VERSION=${VERSION:-latest}

usage() {
    cat << EOF
This script is used to install chaos-mesh.
Before running this script, please ensure that:
* have installed docker if you run chaos-mesh in local.
* have installed Kubernetes if you run chaos-mesh in normal Kubernetes cluster
USAGE:
    install.sh [FLAGS] [OPTIONS]
FLAGS:
    -h, --help               Prints help information
    -d, --dependency-only    Install dependencies only, including kind, kubectl, local-kube.
        --force              Force reinstall all components if they are already installed, include: kind, local-kube, chaos-mesh
        --force-chaos-mesh   Force reinstall chaos-mesh if it is already installed
        --force-local-kube   Force reinstall local Kubernetes cluster if it is already installed
        --force-kubectl      Force reinstall kubectl client if it is already installed
        --force-kind         Force reinstall Kind if it is already installed
        --volume-provisioner Deploy volume provisioner in local Kubernetes cluster
        --local-registry     Deploy local docker registry in local Kubernetes cluster
        --template           Locally render templates
        --k3s                Install chaos-mesh in k3s environment
        --microk8s           Install chaos-mesh in microk8s environment
        --host-network       Install chaos-mesh using hostNetwork
OPTIONS:
    -v, --version            Version of chaos-mesh, default value: ${VERSION}
    -l, --local [kind]       Choose a way to run a local kubernetes cluster, supported value: kind,
                             If this value is not set and the Kubernetes is not installed, this script will exit with 1.
    -n, --name               Name of Kubernetes cluster, default value: kind
    -c  --crd                The path of the crd files. Get the crd file from "https://mirrors.chaos-mesh.org" if the crd path is empty.
    -r  --runtime            Runtime specifies which container runtime to use. Currently we only supports docker and containerd. default value: docker
        --kind-version       Version of the Kind tool, default value: v0.11.1
        --node-num           The count of the cluster nodes,default value: 3
        --k8s-version        Version of the Kubernetes cluster,default value: v1.17.2
        --volume-num         The volumes number of each kubernetes node,default value: 5
        --release-name       Release name of chaos-mesh, default value: chaos-mesh
        --namespace          Namespace of chaos-mesh, default value: chaos-mesh
        --timezone           Specifies timezone to be used by chaos-dashboard, chaos-daemon and controller.
EOF
}

main() {
    local local_kube=""
    local cm_version="${VERSION}"
    local kind_name="kind"
    local kind_version="v0.11.1"
    local node_num=3
    local k8s_version="v1.17.2"
    local volume_num=5
    local release_name="chaos-mesh"
    local namespace="chaos-mesh"
    local timezone="UTC"
    local force_chaos_mesh=false
    local force_local_kube=false
    local force_kubectl=false
    local force_kind=false
    local volume_provisioner=false
    local local_registry=false
    local crd=""
    local runtime="docker"
    local template=false
    local install_dependency_only=false
    local k3s=false
    local microk8s=false
    local host_network=false
    local docker_registry="ghcr.io"

    while [[ $# -gt 0 ]]
    do
        key="$1"
        case "$key" in
            -h|--help)
                usage
                exit 0
                ;;
            -l|--local)
                local_kube="$2"
                shift
                shift
                ;;
            -v|--version)
                cm_version="$2"
                shift
                shift
                ;;
            -n|--name)
                kind_name="$2"
                shift
                shift
                ;;
            -c|--crd)
                crd="$2"
                shift
                shift
                ;;
            -r|--runtime)
                runtime="$2"
                shift
                shift
                ;;
            -d|--dependency-only)
                install_dependency_only=true
                shift
                ;;
            --force)
                force_chaos_mesh=true
                force_local_kube=true
                force_kubectl=true
                force_kind=true
                shift
                ;;
            --force-local-kube)
                force_local_kube=true
                shift
                ;;
            --force-kubectl)
                force_kubectl=true
                shift
                ;;
            --force-kind)
                force_kind=true
                shift
                ;;
            --force-chaos-mesh)
                force_chaos_mesh=true
                shift
                ;;
            --template)
                template=true
                shift
                ;;
            --volume-provisioner)
                volume_provisioner=true
                shift
                ;;
            --local-registry)
                local_registry=true
                shift
                ;;
            --kind-version)
                kind_version="$2"
                shift
                shift
                ;;
            --node-num)
                node_num="$2"
                shift
                shift
                ;;
            --k8s-version)
                k8s_version="$2"
                shift
                shift
                ;;
            --volume-num)
                volume_num="$2"
                shift
                shift
                ;;
            --release-name)
                release_name="$2"
                shift
                shift
                ;;
            --namespace)
                namespace="$2"
                shift
                shift
                ;;
            --k3s)
                k3s=true
                shift
                ;;
            --microk8s)
                microk8s=true
                shift
                ;;
            --host-network)
                host_network=true
                shift
                ;;
            --timezone)
                timezone="$2"
                shift
                shift
                ;;
            --docker-registry)
                docker_registry="$2"
                shift
                shift
                ;;
            *)
                echo "unknown flag or option $key"
                usage
                exit 1
                ;;
        esac
    done

    if [ "${runtime}" != "docker" ] && [ "${runtime}" != "containerd" ]; then
        printf "container runtime %s is not supported\n" "${runtime}"
        exit 1
    fi

    if [ "${local_kube}" != "" ] && [ "${local_kube}" != "kind" ]; then
        printf "local Kubernetes by %s is not supported\n" "${local_kube}"
        exit 1
    fi

    if [ "${local_kube}" == "kind" ]; then
        runtime="containerd"
    fi

    if [ "${k3s}" == "true" ]; then
        runtime="containerd"
    fi

    if [ "${microk8s}" == "true" ]; then
        runtime="containerd"
    fi

    if [ "${crd}" == "" ]; then
        crd="https://mirrors.chaos-mesh.org/${cm_version}/crd.yaml"
    fi
    if $template; then
        ensure gen_crd_manifests "${crd}"
        ensure gen_chaos_mesh_manifests "${runtime}" "${k3s}" "${cm_version}" "${timezone}" "${host_network}" "${docker_registry}" "${microk8s}"
        exit 0
    fi

    need_cmd "sed"
    need_cmd "tr"

    if [ "${local_kube}" == "kind" ]; then
        prepare_env
        install_kubectl "${k8s_version}" ${force_kubectl}

        check_docker
        install_kind "${kind_version}" ${force_kind}
        install_kubernetes_by_kind "${kind_name}" "${k8s_version}" "${node_num}" "${volume_num}" ${force_local_kube} ${volume_provisioner} ${local_registry}
    fi

    if [ "${install_dependency_only}" = true ]; then
        exit 0
    fi

    check_kubernetes
    install_chaos_mesh "${release_name}" "${namespace}" "${local_kube}" ${force_chaos_mesh} "${crd}" "${runtime}" "${k3s}" "${cm_version}" "${timezone}" "${docker_registry}" "${microk8s}"
    ensure_pods_ready "${namespace}" "app.kubernetes.io/component=controller-manager" 100
    ensure_pods_ready "${namespace}" "app.kubernetes.io/component=chaos-daemon" 100
    ensure_pods_ready "${namespace}" "app.kubernetes.io/component=chaos-dashboard" 100
    printf "Chaos Mesh %s is installed successfully\n" "${release_name}"
}

prepare_env() {
    mkdir -p "$HOME/local/bin"
    local set_path="export PATH=$HOME/local/bin:\$PATH"
    local env_file="$HOME/.bash_profile"
    if [[ ! -e "${env_file}" ]]; then
        ensure touch "${env_file}"
    fi
    grep -qF -- "${set_path}" "${env_file}" || echo "${set_path}" >> "${env_file}"
    ensure source "${env_file}"
}

check_kubernetes() {
    need_cmd "kubectl"
    kubectl_err_msg=$(kubectl version --output=yaml 2>&1 1>/dev/null)
    if [ "$kubectl_err_msg" != "" ]; then
        printf "check Kubernetes failed, error: %s\n" "${kubectl_err_msg}"
        exit 1
    fi

    check_kubernetes_version
}

check_kubernetes_version() {
    version_info=$(kubectl version --output=yaml | grep gitVersion | sed 's/.*gitVersion: v\([0-9.]*\).*/\1/g')

    for v in $version_info
    do
        if version_lt "$v" "1.12.0"; then
            printf "Chaos Mesh requires Kubernetes cluster running 1.12 or later\n"
            exit 1
        fi
    done
}

install_kubectl() {
    local kubectl_version=$1
    local force_install=$2

    printf "Install kubectl client\n"

    err_msg=$(kubectl version --client=true --output=yaml 2>&1 1>/dev/null)
    if [ "$err_msg" == "" ]; then
        v=$(kubectl version --client=true --output=yaml | grep gitVersion | sed 's/.*gitVersion: v\([0-9.]*\).*/\1/g')
        target_version=$(echo "${kubectl_version}" | sed s/v//g)
        if version_lt "$v" "${target_version}"; then
            printf "Chaos Mesg requires kubectl version %s or later\n"  "${target_version}"
        else
            printf "kubectl Version %s has been installed\n" "$v"
            if [ "$force_install" != "true" ]; then
                return
            fi
        fi
    fi

    need_cmd "curl"
    local KUBECTL_BIN="${HOME}/local/bin/kubectl"
    local target_os=$(lowercase $(uname))

    ensure curl -Lo /tmp/kubectl https://storage.googleapis.com/kubernetes-release/release/${kubectl_version}/bin/${target_os}/amd64/kubectl
    ensure chmod +x /tmp/kubectl
    ensure mv /tmp/kubectl "${KUBECTL_BIN}"
}


install_kubernetes_by_kind() {
    local cluster_name=$1
    local cluster_version=$2
    local node_num=$3
    local volume_num=$4
    local force_install=$5
    local volume_provisioner=$6
    local local_registry=$7

    printf "Install local Kubernetes %s\n" "${cluster_name}"

    need_cmd "kind"

    work_dir=${HOME}/kind/${cluster_name}
    kubeconfig_path=${work_dir}/config
    data_dir=${work_dir}/data
    clusters=$(kind get clusters)
    cluster_exist=false
    for c in $clusters
    do
        if [ "$c" == "$cluster_name" ]; then
            printf "Kind cluster %s has been installed\n" "${cluster_name}"
            cluster_exist=true
            break
        fi
    done

    if [ "$cluster_exist" == "true" ]; then
        if [ "$force_install" == "true" ]; then
            printf "Delete Kind Kubernetes cluster %s\n" "${cluster_name}"
            kind delete cluster --name="${cluster_name}"
            status=$?
            if [ $status -ne 0 ]; then
                printf "Delete Kind Kubernetes cluster %s failed\n" "${cluster_name}"
                exit 1
            fi
        else
            ensure kind get kubeconfig --name="${cluster_name}" > "${kubeconfig_path}"
            return
        fi
    fi

    ensure mkdir -p "${work_dir}"

    printf "Clean data dir: %s\n" "${data_dir}"
    if [ -d "${data_dir}" ]; then
        ensure rm -rf "${data_dir}"
    fi

    config_file=${work_dir}/kind-config.yaml
    cat <<EOF > "${config_file}"
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
kubeadmConfigPatches:
- |
  kind: ClusterConfiguration
  metadata:
    name: config
  apiServerExtraArgs:
    enable-admission-plugins: NodeRestriction,MutatingAdmissionWebhook,ValidatingAdmissionWebhook
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 5000
    hostPort: 5000
    listenAddress: 127.0.0.1
    protocol: TCP
EOF

    for ((i=0;i<"${node_num}";i++))
    do
        ensure mkdir -p "${data_dir}/worker${i}"
        cat <<EOF >>  "${config_file}"
- role: worker
  extraMounts:
EOF
        for ((k=1;k<="${volume_num}";k++))
        do
            ensure mkdir -p "${data_dir}/worker${i}/vol${k}"
            cat <<EOF  >>  "${config_file}"
  - containerPath: /mnt/disks/vol${k}
    hostPath: ${data_dir}/worker${i}/vol${k}
EOF
        done
    done

    local kind_image="kindest/node:${cluster_version}"

    printf "start to create kubernetes cluster %s" "${cluster_name}"
    ensure kind create cluster --config "${config_file}" --image="${kind_image}" --name="${cluster_name}" --retain -v 1
    ensure kind get kubeconfig --name="${cluster_name}" > "${kubeconfig_path}"
    ensure export KUBECONFIG="${kubeconfig_path}"

    if [ "$volume_provisioner" == "true" ]; then
        deploy_volume_provisioner "${work_dir}"
    fi
}

deploy_volume_provisioner() {
    local data_dir=$1
    local config_file=${data_dir}/local-volume-provisionser.yaml

    volume_provisioner_image="quay.io/external_storage/local-volume-provisioner:v2.3.2"

    cat <<EOF >"${config_file}"
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: "local-storage"
provisioner: "kubernetes.io/no-provisioner"
volumeBindingMode: "WaitForFirstConsumer"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: local-provisioner-config
  namespace: kube-system
data:
  nodeLabelsForPV: |
    - kubernetes.io/hostname
  storageClassMap: |
    local-storage:
      hostDir: /mnt/disks
      mountDir: /mnt/disks
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: local-volume-provisioner
  namespace: kube-system
  labels:
    app: local-volume-provisioner
spec:
  selector:
    matchLabels:
      app: local-volume-provisioner
  template:
    metadata:
      labels:
        app: local-volume-provisioner
    spec:
      serviceAccountName: local-storage-admin
      containers:
        - image: ${volume_provisioner_image}
          name: provisioner
          securityContext:
            privileged: true
          env:
          - name: MY_NODE_NAME
            valueFrom:
              fieldRef:
                fieldPath: spec.nodeName
          - name: MY_NAMESPACE
            valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
          - name: JOB_CONTAINER_IMAGE
            value: "quay.io/external_storage/local-volume-provisioner:v2.3.2"
          resources:
            requests:
              cpu: 100m
              memory: 100Mi
            limits:
              cpu: 100m
              memory: 100Mi
          volumeMounts:
            - mountPath: /etc/provisioner/config
              name: provisioner-config
              readOnly: true
            # mounting /dev in DinD environment would fail
            # - mountPath: /dev
            #   name: provisioner-dev
            - mountPath: /mnt/disks
              name: local-disks
              mountPropagation: "HostToContainer"
      volumes:
        - name: provisioner-config
          configMap:
            name: local-provisioner-config
        # - name: provisioner-dev
        #   hostPath:
        #     path: /dev
        - name: local-disks
          hostPath:
            path: /mnt/disks
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: local-storage-admin
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: local-storage-provisioner-pv-binding
  namespace: kube-system
subjects:
- kind: ServiceAccount
  name: local-storage-admin
  namespace: kube-system
roleRef:
  kind: ClusterRole
  name: system:persistent-volume-provisioner
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: local-storage-provisioner-node-clusterrole
  namespace: kube-system
rules:
- apiGroups: [""]
  resources: ["nodes"]
  verbs: ["get"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: local-storage-provisioner-node-binding
  namespace: kube-system
subjects:
- kind: ServiceAccount
  name: local-storage-admin
  namespace: kube-system
roleRef:
  kind: ClusterRole
  name: local-storage-provisioner-node-clusterrole
  apiGroup: rbac.authorization.k8s.io
EOF
    ensure kubectl apply -f "${config_file}"
}

install_kind() {
    local kind_version=$1
    local force_install=$2

    printf "Install Kind tool\n"

    err_msg=$(kind version 2>&1 1>/dev/null)
    if [ "$err_msg" == "" ]; then
        v=$(kind version | awk '{print $2}' | sed s/v//g)
        target_version=${kind_version//v}
        if version_lt "$v" "${target_version}"; then
            printf "Chaos Mesh requires Kind version %s or later\n" "${target_version}"
        else
            printf "Kind Version %s has been installed\n" "$v"
            if [ "$force_install" != "true" ]; then
                return
            fi
        fi
    fi

    local KIND_BIN="${HOME}/local/bin/kind"
    local target_os=$(lowercase $(uname))
    ensure curl -Lo /tmp/kind https://github.com/kubernetes-sigs/kind/releases/download/"$1"/kind-"${target_os}"-amd64
    ensure chmod +x /tmp/kind
    ensure mv /tmp/kind "$KIND_BIN"
}

install_chaos_mesh() {
    local release_name=$1
    local namespace=$2
    local local_kube=$3
    local force_install=$4
    local crd=$5
    local runtime=$6
    local k3s=$7
    local version=$8
    local timezone=$9
    local docker_registry=${10}
    local microk8s=${11}

    printf "Install Chaos Mesh %s\n" "${release_name}"

    gen_crd_manifests "${crd}" | kubectl create --validate=false -f - || exit 1
    gen_chaos_mesh_manifests "${runtime}" "${k3s}" "${version}" "${timezone}" "${host_network}" "${docker_registry}" "${microk8s}" | kubectl apply -f - || exit 1
}

version_lt() {
    vercomp $1 $2
    if [ $? == 2 ];  then
        return 0
    fi

    return 1
}

vercomp () {
    if [[ $1 == $2 ]]
    then
        return 0
    fi
    local IFS=.
    local i ver1 ver2
    read -ra ver1 <<< "$1"
    read -ra ver2 <<< "$2"
    # fill empty fields in ver1 with zeros
    for ((i=${#ver1[@]}; i<${#ver2[@]}; i++))
    do
        ver1[i]=0
    done
    for ((i=0; i<${#ver1[@]}; i++))
    do
        if [[ -z ${ver2[i]} ]]
        then
            # fill empty fields in ver2 with zeros
            ver2[i]=0
        fi
        if ((10#${ver1[i]} > 10#${ver2[i]}))
        then
            return 1
        fi
        if ((10#${ver1[i]} < 10#${ver2[i]}))
        then
            return 2
        fi
    done
    return 0
}

check_docker() {
    need_cmd "docker"
    docker_err_msg=$(docker version 2>&1 1>/dev/null)
    if [ "$docker_err_msg" != "" ]; then
        printf "check docker failed:\n"
        echo "$docker_err_msg"
        exit 1
    fi
}

say() {
    printf 'install chaos-mesh: %s\n' "$1"
}

err() {
    say "$1" >&2
    exit 1
}

need_cmd() {
    if ! check_cmd "$1"; then
        err "need '$1' (command not found)"
    fi
}

check_cmd() {
    command -v "$1" > /dev/null 2>&1
}

lowercase() {
    echo "$@" | tr "[A-Z]" "[a-z]"
}

# Run a command that should never fail. If the command fails execution
# will immediately terminate with an error showing the failing
# command.
ensure() {
    if ! "$@"; then err "command failed: $*"; fi
}

ensure_pods_ready() {
    local namespace=$1
    local labels=""
    local limit=$3

    if [ "$2" != "" ]; then
        labels="-l $2"
    fi

    count=0
    while [ -n "$(kubectl get pods -n "${namespace}" ${labels} --no-headers | grep -v Running)" ];
    do
        echo "Waiting for pod running" && sleep 10;

        kubectl get pods -n "${namespace}" ${labels} --no-headers | >&2 grep -v Running || true

        ((count=count+1))
        if [ $count -gt $limit ]; then
            printf "Waiting for pod status running timeout\n"
            exit 1
        fi
    done
}

gen_crd_manifests() {
    local crd=$1

    if check_url "$crd"; then
        need_cmd curl
        ensure curl -sSL "$crd"
        return
    fi

    ensure cat "$crd"
}

check_url() {
    local url=$1
    local regex='^(https?|ftp|file)://[-A-Za-z0-9\+&@#/%?=~_|!:,.;]*[-A-Za-z0-9\+&@#/%=~_|]\.[-A-Za-z0-9\+&@#/%?=~_|!:,.;]*[-A-Za-z0-9\+&@#/%=~_|]$'
    if [[ $url =~ $regex ]];then
        return 0
    else
        return 1
    fi
}

gen_chaos_mesh_manifests() {
    local runtime=$1
    local k3s=$2
    local version=$3
    local timezone=$4
    local host_network=$5
    local docker_registry=$6
    local microk8s=$7
    local socketDir="/var/run"
    local socketName="docker.sock"
    if [ "${runtime}" == "containerd" ]; then
        socketDir="/run/containerd"
        socketName="containerd.sock"
    fi

    if [ "${k3s}" == "true" ]; then
        socketDir="/run/k3s/containerd"
        socketName="containerd.sock"
    fi

    if [ "${microk8s}" == "true" ]; then
        socketDir="/var/snap/microk8s/common/run"
        socketName="containerd.sock"
    fi

    need_cmd mktemp
    need_cmd openssl
    need_cmd curl

    K8S_SERVICE="chaos-mesh-controller-manager"
    K8S_NAMESPACE="chaos-mesh"
    VERSION_TAG="${version}"

    IMAGE_REGISTRY_PREFIX="${docker_registry}"
    tmpdir=$(mktemp -d)

    ensure openssl genrsa -out ${tmpdir}/ca.key 2048 > /dev/null 2>&1
    ensure openssl req -x509 -new -nodes -key ${tmpdir}/ca.key -subj "/CN=${K8S_SERVICE}.${K8S_NAMESPACE}.svc" -days 1875 -out ${tmpdir}/ca.crt > /dev/null 2>&1
    ensure openssl genrsa -out ${tmpdir}/server.key 2048 > /dev/null 2>&1

    cat <<EOF > ${tmpdir}/csr.conf
[req]
prompt = no
req_extensions = v3_req
distinguished_name = dn
[dn]
CN = ${K8S_SERVICE}.${K8S_NAMESPACE}.svc
[v3_req]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names
[alt_names]
DNS.1 = ${K8S_SERVICE}
DNS.2 = ${K8S_SERVICE}.${K8S_NAMESPACE}
DNS.3 = ${K8S_SERVICE}.${K8S_NAMESPACE}.svc
EOF

    ensure openssl req -new -key ${tmpdir}/server.key -out ${tmpdir}/server.csr -config ${tmpdir}/csr.conf > /dev/null 2>&1
    ensure openssl x509 -req -in ${tmpdir}/server.csr -CA ${tmpdir}/ca.crt -CAkey ${tmpdir}/ca.key -CAcreateserial -out ${tmpdir}/server.crt -days 1875 -extensions v3_req -extfile ${tmpdir}/csr.conf > /dev/null 2>&1

    TLS_KEY=$(openssl base64 -A -in ${tmpdir}/server.key)
    TLS_CRT=$(openssl base64 -A -in ${tmpdir}/server.crt)
    CA_BUNDLE=$(openssl base64 -A -in ${tmpdir}/ca.crt)

    # chaos-mesh.yaml start
    cat <<EOF
---
apiVersion: v1
kind: Namespace
metadata:
  name: chaos-mesh
---
# Source: chaos-mesh/templates/chaos-daemon-rbac.yaml
kind: ServiceAccount
apiVersion: v1
metadata:
  namespace: "chaos-mesh"
  name: chaos-daemon
  labels:
    app.kubernetes.io/name: chaos-mesh
    app.kubernetes.io/instance: chaos-mesh
    app.kubernetes.io/part-of: chaos-mesh
    app.kubernetes.io/version: ${VERSION_TAG##v}
    app.kubernetes.io/component: chaos-daemon
---
# Source: chaos-mesh/templates/chaos-dashboard-rbac.yaml
# Copyright 2022 Chaos Mesh Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# ServiceAccount for component chaos-dashboard
kind: ServiceAccount
apiVersion: v1
metadata:
  namespace: "chaos-mesh"
  name: chaos-dashboard
  labels:
    app.kubernetes.io/name: chaos-mesh
    app.kubernetes.io/instance: chaos-mesh
    app.kubernetes.io/part-of: chaos-mesh
    app.kubernetes.io/version: ${VERSION_TAG##v}
    app.kubernetes.io/component: chaos-dashboard
---
# Source: chaos-mesh/templates/controller-manager-rbac.yaml
# Copyright 2021 Chaos Mesh Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
kind: ServiceAccount
apiVersion: v1
metadata:
  namespace: "chaos-mesh"
  name: chaos-controller-manager
  labels:
    app.kubernetes.io/name: chaos-mesh
    app.kubernetes.io/instance: chaos-mesh
    app.kubernetes.io/part-of: chaos-mesh
    app.kubernetes.io/version: ${VERSION_TAG##v}
    app.kubernetes.io/component: controller-manager
---
# Source: chaos-mesh/templates/dns-rbac.yaml
# Copyright 2021 Chaos Mesh Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
kind: ServiceAccount
apiVersion: v1
metadata:
  namespace: "chaos-mesh"
  name: chaos-dns-server
  labels:
    app.kubernetes.io/name: chaos-mesh
    app.kubernetes.io/instance: chaos-mesh
    app.kubernetes.io/part-of: chaos-mesh
    app.kubernetes.io/version: ${VERSION_TAG##v}
    app.kubernetes.io/component: dns-server
---
# Source: chaos-mesh/templates/secrets-configuration.yaml
# Copyright 2021 Chaos Mesh Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
kind: Secret
apiVersion: v1
metadata:
  name: chaos-mesh-webhook-certs
  namespace: "chaos-mesh"
  labels:
    app.kubernetes.io/name: chaos-mesh
    app.kubernetes.io/instance: chaos-mesh
    app.kubernetes.io/part-of: chaos-mesh
    app.kubernetes.io/version: ${VERSION_TAG##v}
    app.kubernetes.io/component: webhook-secret
type: Opaque
data:
  ca.crt: "${CA_BUNDLE}"
  tls.crt: "${TLS_CRT}"
  tls.key: "${TLS_KEY}"
---
# Source: chaos-mesh/templates/dns-configmap.yaml
# Copyright 2021 Chaos Mesh Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
apiVersion: v1
kind: ConfigMap
metadata:
  name: dns-server-config
  namespace: "chaos-mesh"
  labels:
    app.kubernetes.io/name: chaos-mesh
    app.kubernetes.io/instance: chaos-mesh
    app.kubernetes.io/part-of: chaos-mesh
    app.kubernetes.io/version: ${VERSION_TAG##v}
    app.kubernetes.io/component: chaos-dns-server
data:
  Corefile: |
    .:5353 {
        errors
        health {
            lameduck 5s
        }
        ready
        k8s_dns_chaos cluster.local in-addr.arpa ip6.arpa {
            pods insecure
            fallthrough in-addr.arpa ip6.arpa
            ttl 30
            grpcport 9288
        }
        forward . /etc/resolv.conf {
            max_concurrent 1000
        }
        cache 30
        loop
        reload
        loadbalance
    }
---
# Source: chaos-mesh/templates/chaos-dashboard-rbac.yaml
# ClusterRole for chaos-dashboard at cluster scope
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: chaos-mesh-chaos-dashboard-cluster-level
  labels:
    app.kubernetes.io/name: chaos-mesh
    app.kubernetes.io/instance: chaos-mesh
    app.kubernetes.io/part-of: chaos-mesh
    app.kubernetes.io/version: ${VERSION_TAG##v}
    app.kubernetes.io/component: chaos-dashboard
rules:
  # chaos-dashboard could list namespace for selector hints
  - apiGroups: [ "" ]
    resources:
      - namespaces
    verbs:
      - get
      - list
      - watch
  # chaos-dashboard use subjectaccessreviews to authorize the requests
  - apiGroups: [ "authorization.k8s.io" ]
    resources:
      - subjectaccessreviews
    verbs:
      - create
---
# Source: chaos-mesh/templates/chaos-dashboard-rbac.yaml
# ClusterRole for chaos-dashboard in target namespace
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: chaos-mesh-chaos-dashboard-target-namespace
  labels:
    app.kubernetes.io/name: chaos-mesh
    app.kubernetes.io/instance: chaos-mesh
    app.kubernetes.io/part-of: chaos-mesh
    app.kubernetes.io/version: ${VERSION_TAG##v}
    app.kubernetes.io/component: chaos-dashboard
rules:
  # chaos dashboard could list pods for selector hints
  - apiGroups:
      - ""
    resources:
      - pods
    verbs:
      - get
      - list
      - watch
  # chaos dashboard could record evnets from chaos experiments
  - apiGroups:
      - ""
    resources:
      - events
    verbs:
      - get
      - list
      - watch
  # chaos dashboard could record and manipulate all the Chaos Mesh resources in target namespace
  - apiGroups: [ "chaos-mesh.org" ]
    resources:
      - "*"
    verbs: [ "*" ]
---
# Source: chaos-mesh/templates/controller-manager-rbac.yaml
# roles
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: chaos-mesh-chaos-controller-manager-target-namespace
  labels:
    app.kubernetes.io/name: chaos-mesh
    app.kubernetes.io/instance: chaos-mesh
    app.kubernetes.io/part-of: chaos-mesh
    app.kubernetes.io/version: ${VERSION_TAG##v}
    app.kubernetes.io/component: controller-manager
rules:
  - apiGroups: [ "" ]
    resources: [ "pods", "configmaps", "secrets"]
    verbs: [ "get", "list", "watch", "delete", "update", "patch" ]
  - apiGroups:
      - ""
    resources:
      - pods
    verbs:
      - "create"
  - apiGroups:
      - ""
    resources:
      - "pods/log"
    verbs:
      - "get"
  - apiGroups:
      - ""
    resources:
      - events
    verbs:
      - patch
      - create
      - watch
      - list
      - get
  - apiGroups: [ "chaos-mesh.org" ]
    resources:
      - "*"
    verbs: [ "*" ]
---
# Source: chaos-mesh/templates/controller-manager-rbac.yaml
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: chaos-mesh-chaos-controller-manager-cluster-level
  labels:
    app.kubernetes.io/name: chaos-mesh
    app.kubernetes.io/instance: chaos-mesh
    app.kubernetes.io/part-of: chaos-mesh
    app.kubernetes.io/version: ${VERSION_TAG##v}
    app.kubernetes.io/component: controller-manager
rules:
  - apiGroups: [ "" ]
    resources:
      - nodes
      - persistentvolumes
      - persistentvolumeclaims
      - namespaces
      - services
    verbs: [ "get", "list", "watch" ]
  - apiGroups: [ "authorization.k8s.io" ]
    resources:
      - subjectaccessreviews
    verbs: [ "create" ]
---
# Source: chaos-mesh/templates/dns-rbac.yaml
# roles
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: chaos-mesh-chaos-dns-server-target-namespace
  labels:
    app.kubernetes.io/name: chaos-mesh
    app.kubernetes.io/instance: chaos-mesh
    app.kubernetes.io/part-of: chaos-mesh
    app.kubernetes.io/version: ${VERSION_TAG##v}
    app.kubernetes.io/component: dns-server
rules:
  - apiGroups: [ "" ]
    resources: [ "pods" ]
    verbs: [ "get", "list", "watch" ]
  - apiGroups: [ "" ]
    resources: [ "configmaps" ]
    verbs: [ "*" ]
  - apiGroups: [ "chaos-mesh.org" ]
    resources:
      - "*"
    verbs: [ "*" ]
---
# Source: chaos-mesh/templates/dns-rbac.yaml
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: chaos-mesh-chaos-dns-server-cluster-level
  labels:
    app.kubernetes.io/name: chaos-mesh
    app.kubernetes.io/instance: chaos-mesh
    app.kubernetes.io/part-of: chaos-mesh
    app.kubernetes.io/version: ${VERSION_TAG##v}
    app.kubernetes.io/component: dns-server
rules:
  - apiGroups: [ "" ]
    resources:
      - namespaces
      - services
      - endpoints
      - pods
    verbs: [ "get", "list", "watch" ]
---
# Source: chaos-mesh/templates/chaos-dashboard-rbac.yaml
# ClusterRoleBinding for chaos-dashboard at cluster scope
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: chaos-mesh-chaos-dashboard-cluster-level
  labels:
    app.kubernetes.io/name: chaos-mesh
    app.kubernetes.io/instance: chaos-mesh
    app.kubernetes.io/part-of: chaos-mesh
    app.kubernetes.io/version: ${VERSION_TAG##v}
    app.kubernetes.io/component: chaos-dashboard
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: chaos-mesh-chaos-dashboard-cluster-level
subjects:
  - kind: ServiceAccount
    name: chaos-dashboard
    namespace: "chaos-mesh"
---
# Source: chaos-mesh/templates/chaos-dashboard-rbac.yaml
# binding ClusterRole to ServiceAccount for componnet chaos dashboard
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: chaos-mesh-chaos-dashboard-target-namespace
  # TODO: notice that the targetNamespace is still defined as .Values.controllerManager.targetNamespace, .Values.targetNamespace would be better.
  namespace: "chaos-mesh"
  labels:
    app.kubernetes.io/name: chaos-mesh
    app.kubernetes.io/instance: chaos-mesh
    app.kubernetes.io/part-of: chaos-mesh
    app.kubernetes.io/version: ${VERSION_TAG##v}
    app.kubernetes.io/component: chaos-dashboard
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: chaos-mesh-chaos-dashboard-target-namespace
subjects:
  - kind: ServiceAccount
    name: chaos-dashboard
    namespace: "chaos-mesh"
---
# Source: chaos-mesh/templates/controller-manager-rbac.yaml
# bindings cluster level
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: chaos-mesh-chaos-controller-manager-cluster-level
  labels:
    app.kubernetes.io/name: chaos-mesh
    app.kubernetes.io/instance: chaos-mesh
    app.kubernetes.io/part-of: chaos-mesh
    app.kubernetes.io/version: ${VERSION_TAG##v}
    app.kubernetes.io/component: controller-manager
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: chaos-mesh-chaos-controller-manager-cluster-level
subjects:
  - kind: ServiceAccount
    name: chaos-controller-manager
    namespace: "chaos-mesh"
---
# Source: chaos-mesh/templates/controller-manager-rbac.yaml
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: chaos-mesh-chaos-controller-manager-target-namespace
  namespace: "chaos-mesh"
  labels:
    app.kubernetes.io/name: chaos-mesh
    app.kubernetes.io/instance: chaos-mesh
    app.kubernetes.io/part-of: chaos-mesh
    app.kubernetes.io/version: ${VERSION_TAG##v}
    app.kubernetes.io/component: controller-manager
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: chaos-mesh-chaos-controller-manager-target-namespace
subjects:
  - kind: ServiceAccount
    name: chaos-controller-manager
    namespace: "chaos-mesh"
---
# Source: chaos-mesh/templates/dns-rbac.yaml
# bindings cluster level
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: chaos-mesh-chaos-dns-server-cluster-level
  labels:
    app.kubernetes.io/name: chaos-mesh
    app.kubernetes.io/instance: chaos-mesh
    app.kubernetes.io/part-of: chaos-mesh
    app.kubernetes.io/version: ${VERSION_TAG##v}
    app.kubernetes.io/component: dns-server
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: chaos-mesh-chaos-dns-server-cluster-level
subjects:
  - kind: ServiceAccount
    name: chaos-dns-server
    namespace: "chaos-mesh"
---
# Source: chaos-mesh/templates/dns-rbac.yaml
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: chaos-mesh-chaos-dns-server-target-namespace
  namespace: 
  labels:
    app.kubernetes.io/name: chaos-mesh
    app.kubernetes.io/instance: chaos-mesh
    app.kubernetes.io/part-of: chaos-mesh
    app.kubernetes.io/version: ${VERSION_TAG##v}
    app.kubernetes.io/component: dns-server
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: chaos-mesh-chaos-dns-server-target-namespace
subjects:
  - kind: ServiceAccount
    name: chaos-dns-server
    namespace: "chaos-mesh"
---
# Source: chaos-mesh/templates/controller-manager-rbac.yaml
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: chaos-mesh-chaos-controller-manager-control-plane
  namespace: "chaos-mesh"
  labels:
    app.kubernetes.io/name: chaos-mesh
    app.kubernetes.io/instance: chaos-mesh
    app.kubernetes.io/part-of: chaos-mesh
    app.kubernetes.io/version: ${VERSION_TAG##v}
    app.kubernetes.io/component: controller-manager
rules:
  - apiGroups: [ "" ]
    resources: [ "services", "endpoints", "secrets" ]
    verbs: [ "get", "list", "watch" ]
  - apiGroups: [ "authorization.k8s.io" ]
    resources:
      - subjectaccessreviews
    verbs: [ "create" ]
  - apiGroups: [ "" ]
    resources: [ "pods/exec" ]
    verbs: [ "create" ]
  - apiGroups: [ "coordination.k8s.io" ]
    resources: [ "leases" ]
    verbs: [ "*" ]
  - apiGroups: [ "" ]
    resources: [ "configmaps" ]
    verbs: [ "*" ]
---
# Source: chaos-mesh/templates/dns-rbac.yaml
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: chaos-mesh-chaos-dns-server-control-plane
  namespace: "chaos-mesh"
  labels:
    app.kubernetes.io/name: chaos-mesh
    app.kubernetes.io/instance: chaos-mesh
    app.kubernetes.io/part-of: chaos-mesh
    app.kubernetes.io/version: ${VERSION_TAG##v}
    app.kubernetes.io/component: dns-server
rules:
  - apiGroups: [ "" ]
    resources: [ "configmaps" ]
    verbs: [ "get", "list" ]
---
# Source: chaos-mesh/templates/controller-manager-rbac.yaml
# binding for control plane namespace
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: chaos-mesh-chaos-controller-manager-control-plane
  namespace: "chaos-mesh"
  labels:
    app.kubernetes.io/name: chaos-mesh
    app.kubernetes.io/instance: chaos-mesh
    app.kubernetes.io/part-of: chaos-mesh
    app.kubernetes.io/version: ${VERSION_TAG##v}
    app.kubernetes.io/component: controller-manager
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: chaos-mesh-chaos-controller-manager-control-plane
subjects:
  - kind: ServiceAccount
    name: chaos-controller-manager
    namespace: "chaos-mesh"
---
# Source: chaos-mesh/templates/dns-rbac.yaml
# binding for control plane namespace
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: chaos-mesh-chaos-dns-server-control-plane
  namespace: "chaos-mesh"
  labels:
    app.kubernetes.io/name: chaos-mesh
    app.kubernetes.io/instance: chaos-mesh
    app.kubernetes.io/part-of: chaos-mesh
    app.kubernetes.io/version: ${VERSION_TAG##v}
    app.kubernetes.io/component: dns-server
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: chaos-mesh-chaos-dns-server-control-plane
subjects:
  - kind: ServiceAccount
    name: chaos-dns-server
    namespace: "chaos-mesh"
---
# Source: chaos-mesh/templates/chaos-daemon-service.yaml
# Copyright 2021 Chaos Mesh Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
apiVersion: v1
kind: Service
metadata:
  namespace: "chaos-mesh"
  name: chaos-daemon
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "31766"
  labels:
    app.kubernetes.io/name: chaos-mesh
    app.kubernetes.io/instance: chaos-mesh
    app.kubernetes.io/part-of: chaos-mesh
    app.kubernetes.io/version: ${VERSION_TAG##v}
    app.kubernetes.io/component: chaos-daemon
spec:
  clusterIP: None
  ports:
    - name: grpc
      port: 31767
      targetPort: grpc
      protocol: TCP
    - name: http
      port: 31766
      targetPort: http
      protocol: TCP
  selector:
    app.kubernetes.io/name: chaos-mesh
    app.kubernetes.io/instance: chaos-mesh
    app.kubernetes.io/component: chaos-daemon
---
# Source: chaos-mesh/templates/chaos-dashboard-deployment.yaml
apiVersion: v1
kind: Service
metadata:
  namespace: "chaos-mesh"
  name: chaos-dashboard
  labels:
    app.kubernetes.io/name: chaos-mesh
    app.kubernetes.io/instance: chaos-mesh
    app.kubernetes.io/component: chaos-dashboard
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "2334"
spec:
  selector:
    app.kubernetes.io/name: chaos-mesh
    app.kubernetes.io/instance: chaos-mesh
    app.kubernetes.io/component: chaos-dashboard
  type: NodePort
  ports:
    - protocol: TCP
      port: 2333
      targetPort: 2333
      name: http
    - protocol: TCP
      port: 2334
      targetPort: 2334
      name: metric
---
# Source: chaos-mesh/templates/controller-manager-service.yaml
# Copyright 2021 Chaos Mesh Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
apiVersion: v1
kind: Service
metadata:
  namespace: "chaos-mesh"
  name: chaos-mesh-controller-manager
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "10080"
  labels:
    app.kubernetes.io/name: chaos-mesh
    app.kubernetes.io/instance: chaos-mesh
    app.kubernetes.io/part-of: chaos-mesh
    app.kubernetes.io/version: ${VERSION_TAG##v}
    app.kubernetes.io/component: controller-manager
spec:
  type: ClusterIP
  ports:
    - port: 443
      targetPort: webhook
      protocol: TCP
      name: webhook
    - port: 10081
      targetPort: pprof
      protocol: TCP
      name: pprof
    - port: 10082
      targetPort: ctrl
      protocol: TCP
      name: ctrl
    - port: 10080
      targetPort: http
      protocol: TCP
      name: http
  selector:
    app.kubernetes.io/name: chaos-mesh
    app.kubernetes.io/instance: chaos-mesh
    app.kubernetes.io/component: controller-manager
---
# Source: chaos-mesh/templates/dns-service.yaml
# Copyright 2021 Chaos Mesh Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
apiVersion: v1
kind: Service
metadata:
  name: chaos-mesh-dns-server
  namespace: "chaos-mesh"
  annotations:
    prometheus.io/port: "9153"
    prometheus.io/scrape: "true"
  labels:
    app.kubernetes.io/name: chaos-mesh
    app.kubernetes.io/instance: chaos-mesh
    app.kubernetes.io/part-of: chaos-mesh
    app.kubernetes.io/version: ${VERSION_TAG##v}
    app.kubernetes.io/component: dns-server
spec:
  selector:
    app.kubernetes.io/name: chaos-mesh
    app.kubernetes.io/instance: chaos-mesh
    app.kubernetes.io/component: chaos-dns-server
  ports:
  - name: dns
    port: 53
    targetPort: 5353
    protocol: UDP
  - name: dns-tcp
    port: 53
    targetPort: 5353
    protocol: TCP
  - name: metrics
    port: 9153
    protocol: TCP
  - name: grpc
    port: 9288
    protocol: TCP
---
# Source: chaos-mesh/templates/chaos-daemon-daemonset.yaml
# Copyright 2021 Chaos Mesh Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

apiVersion: apps/v1
kind: DaemonSet
metadata:
  namespace: "chaos-mesh"
  name: chaos-daemon
  labels:
    app.kubernetes.io/component: chaos-daemon
    app.kubernetes.io/name: chaos-mesh
    app.kubernetes.io/instance: chaos-mesh
    app.kubernetes.io/part-of: chaos-mesh
    app.kubernetes.io/version: ${VERSION_TAG##v}
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: chaos-mesh
      app.kubernetes.io/instance: chaos-mesh
      app.kubernetes.io/component: chaos-daemon
  template:
    metadata:
      labels:
        app.kubernetes.io/name: chaos-mesh
        app.kubernetes.io/instance: chaos-mesh
        app.kubernetes.io/part-of: chaos-mesh
        app.kubernetes.io/version: ${VERSION_TAG##v}
        app.kubernetes.io/component: chaos-daemon
      annotations:
    spec:
      hostNetwork: ${host_network}
      serviceAccountName: chaos-daemon
      hostPID: true
      containers:
        - name: chaos-daemon
          image: ${IMAGE_REGISTRY_PREFIX}/chaos-mesh/chaos-daemon:${VERSION_TAG}
          imagePullPolicy: IfNotPresent
          command:
            - /usr/local/bin/chaos-daemon
            - --runtime
            - ${runtime}
            - --http-port
            - !!str 31766
            - --grpc-port
            - !!str 31767
            - --pprof
            - --runtime-socket-path
            - /host-run/${socketName}
          env:
            - name: TZ
              value: ${timezone}
          securityContext:
            privileged: true
            capabilities:
              add:
                - SYS_PTRACE
          volumeMounts:
            - name: socket-path
              mountPath: /host-run
            - name: sys-path
              mountPath: /host-sys
            - name: lib-modules
              mountPath: /lib/modules
          ports:
            - name: grpc
              containerPort: 31767
            - name: http
              containerPort: 31766
      volumes:
        - name: socket-path
          hostPath: 
            path: ${socketDir}
        - name: sys-path
          hostPath:
            path: /sys
        - name: lib-modules
          hostPath:
            path: /lib/modules
---
# Source: chaos-mesh/templates/chaos-dashboard-deployment.yaml
# Copyright 2021 Chaos Mesh Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: "chaos-mesh"
  name: chaos-dashboard
  labels:
    app.kubernetes.io/name: chaos-mesh
    app.kubernetes.io/instance: chaos-mesh
    app.kubernetes.io/part-of: chaos-mesh
    app.kubernetes.io/version: ${VERSION_TAG##v}
    app.kubernetes.io/component: chaos-dashboard
spec:
  replicas: 1
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app.kubernetes.io/name: chaos-mesh
      app.kubernetes.io/instance: chaos-mesh
      app.kubernetes.io/component: chaos-dashboard
  template:
    metadata:
      labels:
        app.kubernetes.io/name: chaos-mesh
        app.kubernetes.io/instance: chaos-mesh
        app.kubernetes.io/part-of: chaos-mesh
        app.kubernetes.io/version: ${VERSION_TAG##v}
        app.kubernetes.io/component: chaos-dashboard
      annotations:
    spec:
      securityContext:
            {}
      serviceAccountName: chaos-dashboard
      containers:
        - name: chaos-dashboard
          image: ${IMAGE_REGISTRY_PREFIX}/chaos-mesh/chaos-dashboard:${VERSION_TAG}
          imagePullPolicy: IfNotPresent
          resources:
            limits: {}
            requests:
              cpu: 25m
              memory: 256Mi
          command:
            - /usr/local/bin/chaos-dashboard
          env:
            - name: CLEAN_SYNC_PERIOD
              value: "12h"
            - name: DATABASE_DATASOURCE
              value: "/data/core.sqlite"
            - name: DATABASE_DRIVER
              value: "sqlite3"
            - name: LISTEN_HOST
              value: "0.0.0.0"
            - name: LISTEN_PORT
              value: "2333"
            - name: METRIC_HOST
              value: "0.0.0.0"
            - name: METRIC_PORT
              value: "2334"
            - name: TTL_EVENT
              value: "168h"
            - name: TTL_EXPERIMENT
              value: "336h"
            - name: TTL_SCHEDULE
              value: "336h"
            - name: TTL_WORKFLOW
              value: "336h"
            - name: TZ
              value: ${timezone}
            - name: CLUSTER_SCOPED
              value: "true"
            - name: TARGET_NAMESPACE
              value: "chaos-mesh"
            - name: ENABLE_FILTER_NAMESPACE
              value: "false"
            - name: SECURITY_MODE
              value: "false"
            - name: GCP_SECURITY_MODE
              value: "false"
            - name: GCP_CLIENT_ID
              value: ""
            - name: GCP_CLIENT_SECRET
              value: ""
            - name: DNS_SERVER_CREATE
              value: "true"
            - name: ROOT_URL
              value: "http://localhost:2333"
            - name: ENABLE_PROFILING
              value: "true"
          volumeMounts:
            - name: storage-volume
              mountPath: /data
              subPath: ""
          ports:
            - name: http
              containerPort: 2333
            - name: metric
              containerPort: 2334
      volumes:
      - name: storage-volume
        emptyDir: {}
---
# Source: chaos-mesh/templates/controller-manager-deployment.yaml
# Copyright 2021 Chaos Mesh Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: "chaos-mesh"
  name: chaos-controller-manager
  labels:
    app.kubernetes.io/name: chaos-mesh
    app.kubernetes.io/instance: chaos-mesh
    app.kubernetes.io/part-of: chaos-mesh
    app.kubernetes.io/version: ${VERSION_TAG##v}
    app.kubernetes.io/component: controller-manager
spec:
  replicas: 3
  selector:
    matchLabels:
      app.kubernetes.io/name: chaos-mesh
      app.kubernetes.io/instance: chaos-mesh
      app.kubernetes.io/component: controller-manager
  template:
    metadata:
      labels:
        app.kubernetes.io/name: chaos-mesh
        app.kubernetes.io/instance: chaos-mesh
        app.kubernetes.io/part-of: chaos-mesh
        app.kubernetes.io/version: ${VERSION_TAG##v}
        app.kubernetes.io/component: controller-manager
      annotations:
        rollme: "install.sh"
    spec:
      securityContext:
            {}
      hostNetwork: ${host_network}
      serviceAccountName: chaos-controller-manager
      containers:
      - name: chaos-mesh
        image: ${IMAGE_REGISTRY_PREFIX}/chaos-mesh/chaos-mesh:${VERSION_TAG}
        imagePullPolicy: IfNotPresent
        resources:
            limits: {}
            requests:
              cpu: 25m
              memory: 256Mi
        command:
          - /usr/local/bin/chaos-controller-manager
        env:
          - name: METRICS_PORT
            value: "10080"
          - name: WEBHOOK_PORT
            value: "10250"
          - name: NAMESPACE
            valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
          - name: TEMPLATE_NAMESPACE
            valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
          - name: ALLOW_HOST_NETWORK_TESTING
            value: "false"
          - name: TARGET_NAMESPACE
            value: "chaos-mesh"
          - name: CLUSTER_SCOPED
            value: "true"
          - name: TZ
            value: ${timezone}
          - name: CHAOS_DAEMON_SERVICE_PORT
            value: !!str 31767
          - name: BPFKI_PORT
            value: !!str 50051
          - name: ENABLED_CONTROLLERS
            value: "*"
          - name: ENABLED_WEBHOOKS
            value: "*"
          - name: TEMPLATE_LABELS
            value: "app.kubernetes.io/component:template"
          - name: CONFIGMAP_LABELS
            value: "app.kubernetes.io/component:webhook"
          - name: ENABLE_FILTER_NAMESPACE
            value: "false"
          - name: PPROF_ADDR
            value: ":10081"
          - name: CTRL_ADDR
            value: ":10082"
          - name: CHAOS_DNS_SERVICE_NAME
            value: chaos-mesh-dns-server
          - name: CHAOS_DNS_SERVICE_PORT
            value: !!str 9288
          - name: SECURITY_MODE
            value: "false"
          - name: CHAOSD_SECURITY_MODE
            value: "false"
          - name: EXTRA_CA_TRUST_PATH
            value: /etc/extra-ca-trust
          - name: POD_FAILURE_PAUSE_IMAGE
            value: gcr.io/google-containers/pause:latest
          - name: ENABLE_LEADER_ELECTION
            value: "true"
          - name: LEADER_ELECT_LEASE_DURATION
            value: "15s"
          - name: LEADER_ELECT_RENEW_DEADLINE
            value: "10s"
          - name: LEADER_ELECT_RETRY_PERIOD
            value: "2s"
        volumeMounts:
          - name: webhook-certs
            mountPath: /etc/webhook/certs
            readOnly: true
        ports:
          - name: webhook
            containerPort: 10250
          - name: http
            containerPort: 10080
          - name: pprof
            containerPort: 10081
          - name: ctrl
            containerPort: 10082
      volumes:
        - name: webhook-certs
          secret:
            secretName: chaos-mesh-webhook-certs
---
# Source: chaos-mesh/templates/dns-deployment.yaml
# Copyright 2021 Chaos Mesh Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
apiVersion: apps/v1
kind: Deployment
metadata:
  name: chaos-dns-server
  namespace: "chaos-mesh"
  labels:
    app.kubernetes.io/name: chaos-mesh
    app.kubernetes.io/instance: chaos-mesh
    app.kubernetes.io/part-of: chaos-mesh
    app.kubernetes.io/version: ${VERSION_TAG##v}
    app.kubernetes.io/component: chaos-dns-server
spec:
  replicas: 1
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: chaos-mesh
      app.kubernetes.io/instance: chaos-mesh
      app.kubernetes.io/component: chaos-dns-server
  template:
    metadata:
      labels:
        app.kubernetes.io/name: chaos-mesh
        app.kubernetes.io/instance: chaos-mesh
        app.kubernetes.io/part-of: chaos-mesh
        app.kubernetes.io/version: ${VERSION_TAG##v}
        app.kubernetes.io/component: chaos-dns-server
    spec:
      serviceAccountName: chaos-dns-server
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - podAffinityTerm:
              labelSelector:
                matchExpressions:
                - key: app.kubernetes.io/component
                  operator: In
                  values:
                  - chaos-dns-server
              topologyKey: kubernetes.io/hostname
            weight: 100
      priorityClassName: 
      containers:
      - name: chaos-dns-server
        image: ghcr.io/chaos-mesh/chaos-coredns:v0.2.6
        imagePullPolicy: IfNotPresent
        resources:
          limits: {}
          requests:
            cpu: 100m
            memory: 70Mi
        args: [ "-conf", "/etc/chaos-dns/Corefile" ]
        volumeMounts:
        - name: config-volume
          mountPath: /etc/chaos-dns
          readOnly: true
        ports:
        - containerPort: 5353
          name: dns
          protocol: UDP
        - containerPort: 5353
          name: dns-tcp
          protocol: TCP
        - containerPort: 9153
          name: metrics
          protocol: TCP
        - containerPort: 9288
          name: grpc
          protocol: TCP
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
            scheme: HTTP
          initialDelaySeconds: 60
          timeoutSeconds: 5
          successThreshold: 1
          failureThreshold: 5
        readinessProbe:
          httpGet:
            path: /ready
            port: 8181
            scheme: HTTP
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            add:
            - NET_BIND_SERVICE
            drop:
            - all
          readOnlyRootFilesystem: true
      dnsPolicy: Default
      volumes:
        - name: config-volume
          configMap:
            name: dns-server-config
            items:
            - key: Corefile
              path: Corefile
---
# Source: chaos-mesh/templates/cert-manager-certs.yaml
# Copyright 2022 Chaos Mesh Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
---
# Source: chaos-mesh/templates/chaos-daemon-rbac.yaml
# Copyright 2021 Chaos Mesh Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
---
# Source: chaos-mesh/templates/chaos-dashboard-pvc.yaml
# Copyright 2021 Chaos Mesh Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
---
# Source: chaos-mesh/templates/ingress.yaml
# Copyright 2021 Chaos Mesh Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
---
# Source: chaos-mesh/templates/prometheus-configmap.yaml
# Copyright 2021 Chaos Mesh Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
---
# Source: chaos-mesh/templates/prometheus-deployment.yaml
# Copyright 2021 Chaos Mesh Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
---
# Source: chaos-mesh/templates/prometheus-rbac.yaml
# Copyright 2021 Chaos Mesh Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
---
# Source: chaos-mesh/templates/prometheus-service.yaml
# Copyright 2021 Chaos Mesh Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
---
# Source: chaos-mesh/templates/mutating-admission-webhooks.yaml
# Copyright 2022 Chaos Mesh Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

apiVersion: admissionregistration.k8s.io/v1
kind: MutatingWebhookConfiguration
metadata:
  name: chaos-mesh-mutation
  labels:
    app.kubernetes.io/name: chaos-mesh
    app.kubernetes.io/instance: chaos-mesh
    app.kubernetes.io/part-of: chaos-mesh
    app.kubernetes.io/version: ${VERSION_TAG##v}
    app.kubernetes.io/component: admission-webhook
webhooks:
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /mutate-chaos-mesh-org-v1alpha1-podchaos
    failurePolicy: Fail
    name: mpodchaos.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - podchaos
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /mutate-chaos-mesh-org-v1alpha1-iochaos
    failurePolicy: Fail
    name: miochaos.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - iochaos
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /mutate-chaos-mesh-org-v1alpha1-timechaos
    failurePolicy: Fail
    name: mtimechaos.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - timechaos
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /mutate-chaos-mesh-org-v1alpha1-networkchaos
    failurePolicy: Fail
    name: mnetworkchaos.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - networkchaos
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /mutate-chaos-mesh-org-v1alpha1-kernelchaos
    failurePolicy: Fail
    name: mkernelchaos.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - kernelchaos
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /mutate-chaos-mesh-org-v1alpha1-stresschaos
    failurePolicy: Fail
    name: mstresschaos.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - stresschaos
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /mutate-chaos-mesh-org-v1alpha1-awschaos
    failurePolicy: Fail
    name: mawschaos.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - awschaos
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /mutate-chaos-mesh-org-v1alpha1-azurechaos
    failurePolicy: Fail
    name: mazurechaos.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - azurechaos
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /mutate-chaos-mesh-org-v1alpha1-gcpchaos
    failurePolicy: Fail
    name: mgcpchaos.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - gcpchaos
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /mutate-chaos-mesh-org-v1alpha1-dnschaos
    failurePolicy: Fail
    name: mdnschaos.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - dnschaos
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /mutate-chaos-mesh-org-v1alpha1-jvmchaos
    failurePolicy: Fail
    name: mjvmchaos.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - jvmchaos
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /mutate-chaos-mesh-org-v1alpha1-schedule
    failurePolicy: Fail
    name: mschedule.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - schedules
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /mutate-chaos-mesh-org-v1alpha1-workflow
    failurePolicy: Fail
    name: mworkflow.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - workflows
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /mutate-chaos-mesh-org-v1alpha1-httpchaos
    failurePolicy: Fail
    name: mhttpchaos.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - httpchaos
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /mutate-chaos-mesh-org-v1alpha1-blockchaos
    failurePolicy: Fail
    name: mblockchaos.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - blockchaos
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /mutate-chaos-mesh-org-v1alpha1-physicalmachinechaos
    failurePolicy: Fail
    name: mphysicalmachinechaos.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - physicalmachinechaos
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /mutate-chaos-mesh-org-v1alpha1-physicalmachine
    failurePolicy: Fail
    name: mphysicalmachine.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - physicalmachines
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /mutate-chaos-mesh-org-v1alpha1-statuscheck
    failurePolicy: Fail
    name: mstatuscheck.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - statuschecks
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /mutate-chaos-mesh-org-v1alpha1-remotecluster
    failurePolicy: Fail
    name: mremotecluster.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - remotecluster
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /mutate-chaos-mesh-org-v1alpha1-cloudstackvmchaos
    failurePolicy: Fail
    name: mcloudstackvmchaos.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - cloudstackvmchaos
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /mutate-chaos-mesh-org-v1alpha1-ciliumchaos
    failurePolicy: Fail
    name: mciliumchaos.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - ciliumchaos
---
# Source: chaos-mesh/templates/validating-admission-webhooks.yaml
# Copyright 2022 Chaos Mesh Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: chaos-mesh-validation
  labels:
    app.kubernetes.io/name: chaos-mesh
    app.kubernetes.io/instance: chaos-mesh
    app.kubernetes.io/part-of: chaos-mesh
    app.kubernetes.io/version: ${VERSION_TAG##v}
    app.kubernetes.io/component: admission-webhook
webhooks:
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /validate-chaos-mesh-org-v1alpha1-podchaos
    failurePolicy: Fail
    name: vpodchaos.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - podchaos
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /validate-chaos-mesh-org-v1alpha1-iochaos
    failurePolicy: Fail
    name: viochaos.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - iochaos
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /validate-chaos-mesh-org-v1alpha1-timechaos
    failurePolicy: Fail
    name: vtimechaos.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - timechaos
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /validate-chaos-mesh-org-v1alpha1-networkchaos
    failurePolicy: Fail
    name: vnetworkchaos.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - networkchaos
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /validate-chaos-mesh-org-v1alpha1-kernelchaos
    failurePolicy: Fail
    name: vkernelchaos.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - kernelchaos
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /validate-chaos-mesh-org-v1alpha1-stresschaos
    failurePolicy: Fail
    name: vstresschaos.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - stresschaos
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /validate-chaos-mesh-org-v1alpha1-awschaos
    failurePolicy: Fail
    name: vawschaos.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - awschaos
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /validate-chaos-mesh-org-v1alpha1-azurechaos
    failurePolicy: Fail
    name: vazurechaos.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - azurechaos
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /validate-chaos-mesh-org-v1alpha1-gcpchaos
    failurePolicy: Fail
    name: vgcpchaos.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - gcpchaos
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /validate-chaos-mesh-org-v1alpha1-dnschaos
    failurePolicy: Fail
    name: vdnschaos.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - dnschaos
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /validate-chaos-mesh-org-v1alpha1-jvmchaos
    failurePolicy: Fail
    name: vjvmchaos.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - jvmchaos
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /validate-chaos-mesh-org-v1alpha1-schedule
    failurePolicy: Fail
    name: vschedule.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - schedules
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /validate-chaos-mesh-org-v1alpha1-workflow
    failurePolicy: Fail
    name: vworkflow.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - workflows
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /validate-chaos-mesh-org-v1alpha1-httpchaos
    failurePolicy: Fail
    name: vhttpchaos.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - httpchaos
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /validate-chaos-mesh-org-v1alpha1-blockchaos
    failurePolicy: Fail
    name: vblockchaos.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - blockchaos
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /validate-chaos-mesh-org-v1alpha1-physicalmachinechaos
    failurePolicy: Fail
    name: vphysicalmachinechaos.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - physicalmachinechaos
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /validate-chaos-mesh-org-v1alpha1-physicalmachine
    failurePolicy: Fail
    name: vphysicalmachine.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - physicalmachines
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /validate-chaos-mesh-org-v1alpha1-statuscheck
    failurePolicy: Fail
    name: vstatuscheck.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - statuschecks
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /validate-chaos-mesh-org-v1alpha1-remotecluster
    failurePolicy: Fail
    name: vremotecluster.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - remotecluster
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /validate-chaos-mesh-org-v1alpha1-cloudstackvmchaos
    failurePolicy: Fail
    name: vcloudstackvmchaos.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - cloudstackvmchaos
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /validate-chaos-mesh-org-v1alpha1-ciliumchaos
    failurePolicy: Fail
    name: vciliumchaos.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - ciliumchaos
---
# Source: chaos-mesh/templates/validating-admission-webhooks.yaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: chaos-mesh-validation-auth
  labels:
    app.kubernetes.io/name: chaos-mesh
    app.kubernetes.io/instance: chaos-mesh
    app.kubernetes.io/part-of: chaos-mesh
    app.kubernetes.io/version: ${VERSION_TAG##v}
    app.kubernetes.io/component: admission-webhook
webhooks:
  - clientConfig:
      caBundle: "${CA_BUNDLE}"
      service:
        name: chaos-mesh-controller-manager
        namespace: "chaos-mesh"
        path: /validate-auth
    failurePolicy: Fail
    name: vauth.kb.io
    timeoutSeconds: 5
    sideEffects: None
    admissionReviewVersions: ["v1", "v1beta1"]
    rules:
      - apiGroups:
          - chaos-mesh.org
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources: [ "*" ]
EOF
    # chaos-mesh.yaml end
}

main "$@" || exit 1
