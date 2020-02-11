#!/usr/bin/env bash

# This is a script to quickly install chaos-mesh.
# This script will check if docker and kubernetes are installed. If local mode is set and kubernetes is not installed,
# it will use kind or minikube to install the kubernetes cluster according to the configuration.
# Finally, when all dependencies are installed, chaos-mesh will be installed using helm.

set -e

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
OPTIONS:
    -v, --version           Version of chaos-mesh, default value: latest
    -l, --local [kind]      Choose a way to run a local kubernetes cluster, supported value: kind,
                            If this value is not set and the Kubernetes is not installed, this script will exit with 1.
    -n, --name              Name of Kubernetes cluster, default value: kind
        --kind-version      Version of the Kind tool, default value: v0.7.0
        --node-num          The count of the cluster nodes,default value: 6
        --k8s-version       Version of the Kubernetes cluster,default value: v1.12.8
        --volume-num        The volumes number of each kubernetes node,default value: 9
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
    local node_num=6
    local k8s_version="v1.12.8"
    local volume_num=9
    local helm_version="v2.16.1"
    local release_name="chaos-mesh"
    local namespace="chaos-testing"

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
		printf "local Kubernetes by %s is not supported" "${local_kube}"
		exit 1
    fi

	need_cmd "sed"
	need_cmd "tr"
	prepare_env

	install_helm "${helm_version}"

	if [ "${local_kube}" == "" ]; then
		check_kubernetes
	else
		check_docker
		install_kind "${kind_version}"
		install_kubernetes_by_kind "${kind_name}" "${k8s_version}" "${node_num}" "${volume_num}"
	fi

	install_chaos_mesh "${release_name}" "${namespace}" "${local-kube}"
}

prepare_env() {
	mkdir -p ~/local/bin
	export PTAH=~/local/bin:$PATH
}

check_kubernetes() {
	need_cmd "kubectl"
    kubectl_err_msg=$(kubectl version 2>&1 1>/dev/null)
    if [ "$kubectl_err_msg" != "" ]; then
		printf "check Kubernetes failed:\n"
		printf "%s" "$kubectl_err_msg"
        exit 1
    fi

    check_kubernetes_version
}

check_kubernetes_version() {
    version_info=$(kubectl version | sed 's/.*GitVersion:\"v\([0-9.]*\).*/\1/g')

    for v in $version_info
    do
		if version_lt "$v" "1.12.0"; then
			printf "Chaos Mesh requires Kubernetes cluster running 1.12 or later"
			exit 1
		fi
    done
}

install_kubernetes_by_kind() {
	echo "install kubernetes"
	local cluster_name=$1
	local cluster_version=$2
	local node_num=$3
	local volume_num=$4

	work_dir=${HOME}/kind/${cluster_name}
	kubeconfig_path=${work_dir}/config
	data_dir=${work_dir}/data
	clusters=$(kind get clusters)
	cluster_exist=false
	for c in $clusters
	do
		if [ "$c" == "$cluster_name" ]; then
			printf "Kind cluster %s had been installed" "$cluster_name"
			cluster_exist=true
			break
		fi
	done

	if ! $cluster_exist; then
		mkdir -p "${work_dir}"

		echo "clean data dir: ${data_dir}"
		if [ -d "${data_dir}" ]; then
			rm -rf "${data_dir}"
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
			mkdir -p "${data_dir}/worker${i}"
			cat <<EOF >>  "${config_file}"
- role: worker
  extraMounts:
EOF
			for ((k=1;k<="${volume_num}";k++))
			do
				mkdir -p "${data_dir}/worker${i}/vol${k}"
				cat <<EOF  >>  "${config_file}"
  - containerPath: /mnt/disks/vol${k}
    hostPath: ${data_dir}/worker${i}/vol${k}
EOF
			done
		done

		printf "start to create kubernetes cluster %s" "${cluster_name}"
		kind create cluster --config "${config_file}" --image kindest/node:${cluster_version} --name=${cluster_name}
		kind get kubeconfig --name="${cluster_name}" > "${kubeconfig_path}"
		export KUBECONFIG="${kubeconfig_path}"

		deploy_registry "${cluster_name}" "${data_dir}"
		init_helm "${data_dir}"
	fi
}

deploy_registry() {
	local cluster_name=$1
	local data_dir=$2
	echo "deploy docker registry in kind"
	registry_node=${cluster_name}-control-plane
	registry_node_ip=$(kubectl get nodes "${registry_node}" -o template --template='{{range.status.addresses}}{{if eq .type "InternalIP"}}{{.address}}{{end}}{{end}}')
	registry_file=${data_dir}/registry.yaml

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
        image: registry:2
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
          image: alpine/socat:1.0.5
          args:
          - tcp-listen:5000,fork,reuseaddr
          - tcp-connect:${registry_node_ip}:5000
EOF
	kubectl apply -f "${registry_file}"
}

deploy_volume_provisioner() {
	local data_dir=$1
	local config_file=${data_dir}/local-volume-provisionser.yaml
	local config_url="https://raw.githubusercontent.com/pingcap/chaos-mesh/master/manifests/local-volume-provisioner.yaml"

	rm -rf "${config_file}"
	wget -O "${config_file}" "$config_url"
	kubectl apply -f "${config_file}"
}

install_kind() {
	echo "install kind"

	err_msg=$(kind version 2>&1 1>/dev/null)
	if [ "$err_msg" == "" ]; then
		v=$(kind version | awk '{print $2}' | sed s/v//g)
		target_version=$(echo "$1" | sed s/v//g)
		if version_lt "$v" "${target_version}"; then
			printf "Chaos Mesh requires Kind version %s or later\n" "${target_version}"
		else
			printf "Kind Version %s had installed\n" "$v"
			return
		fi
	fi

	local KIND_BIN="${HOME}/local/bin/kind"
	local target_os=$(lowercase $(uname))
	curl -Lo /tmp/kind https://github.com/kubernetes-sigs/kind/releases/download/"$1"/kind-"${target_os}"-amd64
	chmod +x /tmp/kind
	mv /tmp/kind "$KIND_BIN"
}

install_helm() {
	echo "install helm"

	err_msg=$(helm version --client 2>&1 1>/dev/null)
	if [ "$err_msg" == "" ]; then
		v=$(helm version --client | sed 's/.*SemVer:\"v\([0-9.]*\).*/\1/g')
		target_version=$(echo "$1" | sed s/v//g)
		if version_lt "$v" "${target_version}"; then
			printf "Chaos Mesh requires Helm version %s or later\n" "${target_version}"
		else
			printf "Helm Version %s had installed\n" "$v"
			return
		fi
	fi

	need_cmd "wget"
	need_cmd "tar"

	local HELM_BIN="${HOME}/local/bin/helm"
	local target_os=$(lowercase $(uname))
	local TAR_NAME="helm-$1-$target_os-amd64.tar.gz"
	rm -rf "${TAR_NAME}"
	wget "https://get.helm.sh/${TAR_NAME}"

	tar zxvf "${TAR_NAME}"
	mv "${target_os}"-amd64/helm ${HELM_BIN}
	chmod +x "${HELM_BIN}"
	rm -rf "${TAR_NAME}" "${target_os}"-amd64
}

init_helm() {
	local data_dir=$1
	local rbac_config=${data_dir}/tiller-rbac.yaml
	local rbac_config_url="https://raw.githubusercontent.com/pingcap/chaos-mesh/master/manifests/tiller-rbac.yaml"

	rm -rf "${rbac_config}"
	wget -O "${rbac_config}" "$rbac_config_url"
	kubectl apply -f "${rbac_config}"

	if [[ $(helm version --client --short) == "Client: v2"* ]]; then helm init --service-account=tiller --wait; fi
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
	echo "install chaos mesh"

	local release_name=$1
	local namespace=$2
	local local_kube=$3
	if check_chaos_mesh_installed "${release_name}"; then
		printf "Chaos Mesh %s had been installed" "${release_name}"
		return
	fi

	kubectl apply -f manifests/crd.yaml

	ns_err_msg=$(kubectl get ns "$namespace" 2>&1 1>/dev/null)
	if [ "$ns_err_msg" != "" ]; then
		kubectl create ns chaos-testing
	fi

	if [[ $(helm version --client --short) == "Client: v2"* ]]; then
		if [ "${local_kube}" == "kind" ]; then
			helm install helm/chaos-mesh --name="${release_name}" --namespace="${namespace}" --set chaosDaemon.runtime=containerd --set chaosDaemon.socketPath=/run/containerd/containerd.sock
		else
			helm install helm/chaos-mesh --name="${release_name}" --namespace="${namespace}"
		fi
	else
		if [ "${local_kube}" == "kind" ]; then
			helm install "${release_name}" helm/chaos-mesh --namespace="${namespace}" --set chaosDaemon.runtime=containerd --set chaosDaemon.socketPath=/run/containerd/containerd.sock
		else
			helm install "${release_name}" helm/chaos-mesh --namespace="${namespace}"
		fi
	fi

	printf "Chaos Mesh %s installed successfully" "${release_name}"
}

function version_le() {
	test "$(echo "$@" | tr " " "\n" | sort -V | head -n 1)" == "$1";
}

function version_lt() {
	test "$(echo "$@" | tr " " "\n" | sort -rV | head -n 1)" != "$1";
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

main "$@" || exit 1
