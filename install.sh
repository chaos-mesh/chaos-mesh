#!/usr/bin/env bash

# This is a script to quickly install chaos-mesh.
# This script will check if docker and kubernetes are installed. If local mode is set and kubernetes is not installed,
# it will use kind or minikube to install the kubernetes cluster according to the configuration.
# Finally, when all dependencies are installed, chaos-mesh will be installed using helm.

usage() {
    cat << EOF
This script is used to install chaos-mesh.
Before running this script, please ensure that:
* have installed docker if you run chaos-mesh in local.
* have installed Kubernetes if you run chaos-mesh in normal Kubernetes cluster
USAGE:
    install.sh [FLAGS] [OPTIONS]
FLAGS:
    -h, --help              Prints help information
        --force             Force reinstall all components if they are already installed, include: helm, kind, local-kube, chaos-mesh
        --force-chaos-mesh  Force reinstall chaos-mesh if it is already installed
        --force-local-kube  Force reinstall local Kubernetes cluster if it is already installed
        --force-kubectl     Force reinstall kubectl client if it is already installed
        --force-kind        Force reinstall Kind if it is already installed
        --force-helm        Force reinstall Helm if it is already installed
        --dashboard         Install Chaos Dashboard
        --docker-mirror     Use docker mirror to pull image, dockerhub.azk8s.cn => docker.io, gcr.azk8s.cn => gcr.io
OPTIONS:
    -v, --version           Version of chaos-mesh, default value: latest
    -l, --local [kind]      Choose a way to run a local kubernetes cluster, supported value: kind,
                            If this value is not set and the Kubernetes is not installed, this script will exit with 1.
    -n, --name              Name of Kubernetes cluster, default value: kind
        --kind-version      Version of the Kind tool, default value: v0.7.0
        --node-num          The count of the cluster nodes,default value: 5
        --k8s-version       Version of the Kubernetes cluster,default value: v1.12.8
        --volume-num        The volumes number of each kubernetes node,default value: 5
        --helm-version      Version of the helm tool, default value: v2.16.1
        --release-name      Release name of chaos-mesh, default value: chaos-mesh
        --namespace         Namespace of chaos-mesh, default value: chaos-testing
EOF
}

main() {
    local local_kube=""
    local cm_version="latest"
    local kind_name="kind"
    local kind_version="v0.7.0"
    local node_num=5
    local k8s_version="v1.12.8"
    local volume_num=5
    local helm_version="v2.16.1"
    local release_name="chaos-mesh"
    local namespace="chaos-testing"
    local force_chaos_mesh=false
    local force_local_kube=false
    local force_kubectl=false
    local force_kind=false
    local force_helm=false
    local install_dashboard=false
    local docker_mirror=false

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
            --force)
                force_chaos_mesh=true
                force_local_kube=true
                force_kubectl=true
                force_kind=true
                force_helm=true
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
            --force-helm)
                force_helm=true
                shift
                ;;
            --force-chaos-mesh)
                force_chaos_mesh=true
                shift
                ;;
            --dashboard)
                install_dashboard=true
                shift
                ;;
            --docker-mirror)
                docker_mirror=true
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
            --helm-version)
                helm_version="$2"
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
            *)
                echo "unknown flag or option $key"
                usage
                exit 1
                ;;
        esac
    done

    if [ "${local_kube}" != "" ] && [ "${local_kube}" != "kind" ]; then
        printf "local Kubernetes by %s is not supported\n" "${local_kube}"
        exit 1
    fi

    need_cmd "sed"
    need_cmd "tr"
    prepare_env

    install_helm "${helm_version}" ${force_helm}
    install_kubectl "${k8s_version}" ${force_kubectl}

    if [ "${local_kube}" == "" ]; then
        check_kubernetes
    else
        check_docker
        install_kind "${kind_version}" ${force_kind}
        install_kubernetes_by_kind "${kind_name}" "${k8s_version}" "${node_num}" "${volume_num}" "${helm_version}" ${force_local_kube} ${docker_mirror}
    fi

    install_chaos_mesh "${release_name}" "${namespace}" "${local_kube}" ${force_chaos_mesh} ${install_dashboard} ${docker_mirror}

    if ! timeout 600 ensure_pods_ready "${namespace}"; then echo "Waiting for pod status running timeout"; fi
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
    kubectl_err_msg=$(kubectl version 2>&1 1>/dev/null)
    if [ "$kubectl_err_msg" != "" ]; then
        printf "check Kubernetes failed, error: %s\n" "${kubectl_err_msg}"
        exit 1
    fi

    check_kubernetes_version
}

check_kubernetes_version() {
    version_info=$(kubectl version | sed 's/.*GitVersion:\"v\([0-9.]*\).*/\1/g')

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

    err_msg=$(kubectl version --client=true 2>&1 1>/dev/null)
    if [ "$err_msg" == "" ]; then
        v=$(kubectl version --client=true | sed 's/.*GitVersion:\"v\([0-9.]*\).*/\1/g')
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
    local helm_version=$5
    local force_install=$6
    local docker_mirror=$7

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
apiVersion: kind.sigs.k8s.io/v1alpha3
kubeadmConfigPatches:
- |
  apiVersion: kubeadm.k8s.io/v1alpha3
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
    if [ "$docker_mirror" == "true" ]; then
        azk8spull "${kind_image}" || true
    fi

    printf "start to create kubernetes cluster %s" "${cluster_name}"
    ensure kind create cluster --config "${config_file}" --image="${kind_image}" --name="${cluster_name}"
    ensure kind get kubeconfig --name="${cluster_name}" > "${kubeconfig_path}"
    ensure export KUBECONFIG="${kubeconfig_path}"

    deploy_volume_provisioner "${work_dir}"
    deploy_registry "${cluster_name}" "${work_dir}" ${docker_mirror}
    init_helm "${work_dir}" "${helm_version}" ${docker_mirror}
}

deploy_registry() {
    local cluster_name=$1
    local data_dir=$2
    local docker_mirror=$3

    printf "Deploy docker registry in kind\n"

    need_cmd "kubectl"

    registry_node=${cluster_name}-control-plane
    registry_node_ip=$(kubectl get nodes "${registry_node}" -o template --template='{{range.status.addresses}}{{if eq .type "InternalIP"}}{{.address}}{{end}}{{end}}')
    registry_file=${data_dir}/registry.yaml

    registry_image="registry:2"
    socat_image="alpine/socat:1.0.5"
    if [ "$docker_mirror" == "true" ]; then
        azk8spull ${registry_image} || true
        kind load docker-image ${registry_image} > /dev/null 2>&1 || true
        azk8spull ${socat_image} || true
        kind load docker-image ${socat_image} > /dev/null 2>&1 || true
    fi

    cat <<EOF >"${registry_file}"
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: registry
spec:
  selector:
    matchLabels:
      app: registry
  template:
    metadata:
      labels:
        app: registry
    spec:
      hostNetwork: true
      nodeSelector:
        kubernetes.io/hostname: ${registry_node}
      tolerations:
      - key: node-role.kubernetes.io/master
        operator: "Equal"
        effect: "NoSchedule"
      containers:
      - name: registry
        image: ${registry_image}
        volumeMounts:
        - name: data
          mountPath: /data
      volumes:
      - name: data
        hostPath:
          path: /data
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: registry-proxy
  labels:
    app: registry-proxy
spec:
  selector:
    matchLabels:
      app: registry-proxy
  template:
    metadata:
      labels:
        app: registry-proxy
    spec:
      hostNetwork: true
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: kubernetes.io/hostname
                operator: NotIn
                values:
                  - ${registry_node}
      tolerations:
      - key: node-role.kubernetes.io/master
        operator: "Equal"
        effect: "NoSchedule"
      containers:
        - name: socat
          image: ${socat_image}
          args:
          - tcp-listen:5000,fork,reuseaddr
          - tcp-connect:${registry_node_ip}:5000
EOF
    ensure kubectl apply -f "${registry_file}"
}

deploy_volume_provisioner() {
    local data_dir=$1
    local config_file=${data_dir}/local-volume-provisionser.yaml

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
        - image: "quay.io/external_storage/local-volume-provisioner:v2.3.2"
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
        target_version=$(echo "${kind_version}" | sed s/v//g)
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

install_helm() {
    local helm_version=$1
    local force_install=$2

    printf "Install Helm tool\n"

    err_msg=$(helm version --client 2>&1 1>/dev/null)
    if [ "$err_msg" == "" ]; then
        v=$(helm version --client | sed 's/.*SemVer:\"v\([0-9.]*\).*/\1/g')
        target_version=$(echo "${helm_version}" | sed s/v//g)
        if version_lt "$v" "${target_version}"; then
            printf "Chaos Mesh requires Helm version %s or later\n" "${target_version}"
        else
            printf "Helm Version %s has been installed\n" "$v"
            if [ "$force_install" != "true" ]; then
                return
            fi
        fi
    fi

    need_cmd "tar"

    local HELM_BIN="${HOME}/local/bin/helm"
    local target_os=$(lowercase $(uname))
    local TAR_NAME="helm-$1-$target_os-amd64.tar.gz"
    rm -rf "${TAR_NAME}"

    ensure $(curl -sL "https://get.helm.sh/${TAR_NAME}" | tar xz)

    ensure mv "${target_os}"-amd64/helm "${HELM_BIN}"
    ensure chmod +x "${HELM_BIN}"
    rm -rf "${target_os}"-amd64
}

init_helm() {
    local data_dir=$1
    local helm_version=$2
    local docker_mirror=$3
    local rbac_config=${data_dir}/tiller-rbac.yaml

    need_cmd "helm"

    cat <<EOF > "${rbac_config}"
apiVersion: v1
kind: ServiceAccount
metadata:
  name: tiller
  namespace: kube-system
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: tiller-clusterrolebinding
subjects:
  - kind: ServiceAccount
    name: tiller
    namespace: kube-system
roleRef:
  kind: ClusterRole
  name: cluster-admin
  apiGroup: ""
EOF

    ensure kubectl apply -f "${rbac_config}"

    local tiller_image="gcr.io/kubernetes-helm/tiller:${helm_version}"
    if [ "$docker_mirror" == "true" ]; then
        azk8spull "${tiller_image}" || true
        kind load docker-image "${tiller_image}" > /dev/null 2>&1 || true
    fi

    if [[ $(helm version --client --short) == "Client: v2"* ]]; then
        ensure helm init --service-account=tiller --tiller-image="${tiller_image}" --wait

        if ! timeout 600 ensure_pods_ready "kube-system"; then echo "Waiting for pod status running timeout"; fi
    fi
}

check_chaos_mesh_installed() {
    local release_name=$1

    err_msg=$(helm get "${release_name}" 2>&1 1>/dev/null)
    if [ "$err_msg" == "" ]; then
        return 0
    fi

    return 1
}

install_chaos_mesh() {
    local release_name=$1
    local namespace=$2
    local local_kube=$3
    local force_install=$4
    local install_dashboard=$5
    local docker_mirror=$6

    printf "Install Chaos Mesh %s\n" "${release_name}"

    if check_chaos_mesh_installed "${release_name}"; then
        printf "Chaos Mesh %s has been installed\n" "${release_name}"

        if [ "$force_install" != "true" ]; then
            return
        fi

        printf "Delete Chaos Mesh %s\n"  "${release_name}"
        err_msg=$(helm delete --purge "${release_name}" 2>&1 1>/dev/null)
        if [ "$err_msg" != "" ] && [[ "$err_msg" != *"not found" ]]; then
            printf "Delete Chaos Mesh %s failed, error: %s\n" "${release_name}" "${err_msg}"
            exit 1
        fi
    fi

    kubectl apply -f manifests/crd.yaml

    ns_err_msg=$(kubectl get ns "$namespace" 2>&1 1>/dev/null)
    if [ "$ns_err_msg" != "" ]; then
        ensure kubectl create ns chaos-testing
    fi

    local dashboard_cmd=""
    if [ "$install_dashboard" == "true" ]; then
        dashboard_cmd="--set dashboard.create=true"
    fi

    local runtime_cmd=""
    if [ "${local_kube}" == "kind" ]; then
        runtime_cmd="--set chaosDaemon.runtime=containerd --set chaosDaemon.socketPath=/run/containerd/containerd.sock"
    fi

    local chaos_mesh_image="pingcap/chaos-mesh:latest"
    local chaos_dashboard_image="pingcap/chaos-dashboard:latest"
    local chaos_daemon_image="pingcap/chaos-daemon:latest"
    local kubectl_image="bitnami/kubectl:latest"

    if [ "$docker_mirror" == "true" ]; then
        azk8spull "${chaos_mesh_image}" || true
        kind load docker-image "${chaos_mesh_image}" > /dev/null 2>&1 || true

        azk8spull "${chaos_dashboard_image}" || true
        kind load docker-image "${chaos_dashboard_image}" > /dev/null 2>&1 || true

        azk8spull "${chaos_daemon_image}" || true
        kind load docker-image "${chaos_daemon_image}" > /dev/null 2>&1 || true

        azk8spull "${kubectl_image}" || true
        kind load docker-image "${kubectl_image}" > /dev/null 2>&1 || true
    fi

    if [[ $(helm version --client --short) == "Client: v2"* ]]; then
        ensure helm install helm/chaos-mesh --name="${release_name}" --namespace="${namespace}" ${runtime_cmd} ${dashboard_cmd}
    else
        ensure helm install "${release_name}" helm/chaos-mesh --namespace="${namespace}" ${runtime_cmd} ${dashboard_cmd}
    fi

    printf "Chaos Mesh %s is installed successfully\n" "${release_name}"
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
    local i ver1=($1) ver2=($2)
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
    while [[ $(kubectl get pods -n "${namespace}" -l app.kubernetes.io/instance=chaos-mesh -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}') != "True" ]];
    do
        echo "waiting for pod running" && sleep 5;
    done
}

timeout() {
    perl -e 'alarm shift; exec @ARGV' "$@";
}

azk8spull() {
	image=$1
	if [ -z $image ]; then
		echo "## azk8spull image name cannot be null."
	else
		array=(`echo $image | tr '/' ' '` )

		domainName=""
		repoName=""
		imageName=""

		if [ ${#array[*]} -eq 3 ]; then
			repoName=${array[1]}
			imageName=${array[2]}
			if [ "${array[0]}"x = "docker.io"x ]; then
				domainName="dockerhub.azk8s.cn"
			elif [ "${array[0]}"x = "gcr.io"x ]; then
				domainName="gcr.azk8s.cn"
			elif [ "${array[0]}"x = "quay.io"x ]; then
				domainName="quay.azk8s.cn"
			else
				echo '## azk8spull can not support pulling $image right now.'
			fi
		elif [ ${#array[*]} -eq 2 ]; then
			if [ "${array[0]}"x = "k8s.gcr.io"x ]; then
				domainName="gcr.azk8s.cn"
				repoName="google_containers"
				imageName=${array[1]}
			else
				domainName="dockerhub.azk8s.cn"
				repoName=${array[0]}
				imageName=${array[1]}
			fi
		elif [ ${#array[*]} -eq 1 ]; then
				domainName="dockerhub.azk8s.cn"
				repoName="library"
				imageName=${array[0]}
		else
			echo '## azk8spull can not support pulling $image right now.'
		fi
		if [ $domainName != "" ]; then
			echo "## azk8spull try to pull image from mirror $domainName/$repoName/$imageName."
			docker pull  $domainName/$repoName/$imageName
			if [ $? -eq 0 ]; then
				echo "## azk8spull try to tag $domainName/$repoName/$imageName to $image."
				docker tag $domainName/$repoName/$imageName $image
				if [ $? -eq 0 ]; then
					echo '## azk8spull finish pulling. '
				fi
			fi
		fi
	fi
}

main "$@" || exit 1
