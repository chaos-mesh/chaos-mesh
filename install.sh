#!/usr/bin/env bash

# This is a script to quickly install chaos-mesh.
# This script will check if docker and kubernetes are installed. If local mode is set and kubernetes is not installed,
# it will use kind or minikube to install the kubernetes cluster according to the configuration.
# Finally, when all dependencies are installed, chaos-mesh will be installed using helm.

set -eu

usage() {
    cat << EOF
This script is used to install chaos-mesh.
Before run this script, please ensure that:
* have installed docker
* have installed kubernetes if you are not run chaos-mesh in locally.
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
        --helm-version      Version of the helm tool, default value: v2.0.0
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
    local helm_version="v2.0.0"

    for arg in "$@"; do
        case "$arg" in
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
            *)
                echo "unknown flag or option $arg"
                usage
                exit
                ;;
        esac
    done

    echo "${local_kube}"
    echo "${cm_version}"
    echo "${kind_name}"
}

check_helm() {
    echo "check helm"
}

check_kind() {
    check_docker
}

check_docker() {
    need_cmd docker
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

main "$@" || exit 1
