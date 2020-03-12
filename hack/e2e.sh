#!/usr/bin/env bash

# Copyright 2020 PingCAP, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# See the License for the specific language governing permissions and
# limitations under the License.

#
# E2E entrypoint script.
#

set -o errexit
set -o nounset
set -o pipefail

ROOT=$(unset CDPATH && cd $(dirname "${BASH_SOURCE[0]}")/.. && pwd)
cd $ROOT

source "${ROOT}/hack/lib.sh"

function usage() {
    cat <<'EOF'
This script is entrypoint to run e2e tests.
Usage: hack/e2e.sh [-h] -- [extra test args]
    -h      show this message and exit
Environments:
    DOCKER_REGISTRY_PREFIX      image docker registry
    IMAGE_TAG                   image tag
    SKIP_BUILD                  skip building binaries
    SKIP_IMAGE_BUILD            skip build and push images
    SKIP_UP                     skip starting the cluster
    SKIP_DOWN                   skip shutting down the cluster
    REUSE_CLUSTER               reuse existing cluster if found
    KUBE_VERSION                the version of Kubernetes to test against
    KUBE_WORKERS                the number of worker nodes (excludes master nodes), defaults: 3
    DOCKER_IO_MIRROR            configure mirror for docker.io
    GCR_IO_MIRROR               configure mirror for gcr.io
    QUAY_IO_MIRROR              configure mirror for quay.io
    KIND_DATA_HOSTPATH          (for kind) the host path of data directory for kind cluster, defaults: none
    GINKGO_NODES                ginkgo nodes to run specs, defaults: 1
    GINKGO_PARALLEL             if set to `y`, will run specs in parallel, the number of nodes will be the number of cpus
    GINKGO_NO_COLOR             if set to `y`, suppress color output in default reporter
Examples:
0) view help
    ./hack/e2e.sh -h
1) run all specs
    ./hack/e2e.sh
    GINKGO_NODES=8 ./hack/e2e.sh # in parallel
2) limit specs to run
    ./hack/e2e.sh -- --ginkgo.focus='Basic'
    ./hack/e2e.sh -- --ginkgo.focus='Backup\sand\srestore'
    See https://onsi.github.io/ginkgo/ for more ginkgo options.
3) reuse the cluster and don't tear down it after the testing
    REUSE_CLUSTER=y SKIP_DOWN=y ./hack/e2e.sh -- <e2e args>
4) use registry mirrors
    DOCKER_IO_MIRROR=https://dockerhub.azk8s.cn QUAY_IO_MIRROR=https://quay.azk8s.cn GCR_IO_MIRROR=https://gcr.azk8s.cn ./hack/e2e.sh -- <e2e args>
EOF

}

while getopts "h?" opt; do
    case "$opt" in
    h|\?)
        usage
        exit 0
        ;;
    esac
done

if [ "${1:-}" == "--" ]; then
    shift
fi

hack::ensure_kind
hack::ensure_kubectl
hack::ensure_helm

DOCKER_REGISTRY=${DOCKER_REGISTRY:-localhost:5000}
IMAGE_TAG=${IMAGE_TAG:-latest}
CLUSTER=${CLUSTER:-chaos-mesh}
KUBECONFIG=${KUBECONFIG:-~/.kube/config}
KUBECONTEXT=kind-$CLUSTER
SKIP_BUILD=${SKIP_BUILD:-}
SKIP_IMAGE_BUILD=${SKIP_IMAGE_BUILD:-}
SKIP_UP=${SKIP_UP:-}
SKIP_DOWN=${SKIP_DOWN:-}
REUSE_CLUSTER=${REUSE_CLUSTER:-}
KIND_DATA_HOSTPATH=${KIND_DATA_HOSTPATH:-none}
KUBE_VERSION=${KUBE_VERSION:-v1.12.10}
KUBE_WORKERS=${KUBE_WORKERS:-3}
DOCKER_IO_MIRROR=${DOCKER_IO_MIRROR:-}
GCR_IO_MIRROR=${GCR_IO_MIRROR:-}
QUAY_IO_MIRROR=${QUAY_IO_MIRROR:-}

echo "DOCKER_REGISTRY: $DOCKER_REGISTRY"
echo "IMAGE_TAG: $IMAGE_TAG"
echo "CLUSTER: $CLUSTER"
echo "KUBECONFIG: $KUBECONFIG"
echo "KUBECONTEXT: $KUBECONTEXT"
echo "SKIP_BUILD: $SKIP_BUILD"
echo "SKIP_IMAGE_BUILD: $SKIP_IMAGE_BUILD"
echo "SKIP_UP: $SKIP_UP"
echo "SKIP_DOWN: $SKIP_DOWN"
echo "KIND_DATA_HOSTPATH: $KIND_DATA_HOSTPATH"
echo "KUBE_VERSION: $KUBE_VERSION"
echo "DOCKER_IO_MIRROR: $DOCKER_IO_MIRROR"
echo "GCR_IO_MIRROR: $GCR_IO_MIRROR"
echo "QUAY_IO_MIRROR: $QUAY_IO_MIRROR"

# https://github.com/kubernetes-sigs/kind/releases/tag/v0.6.1
declare -A kind_node_images
kind_node_images["v1.11.10"]="kindest/node:v1.11.10@sha256:8ebe805201da0a988ee9bbcc2de2ac0031f9264ac24cf2a598774f1e7b324fe1"
kind_node_images["v1.12.10"]="kindest/node:v1.12.10@sha256:c5aeca1433e3230e6c1a96b5e1cd79c90139fd80242189b370a3248a05d77118"
kind_node_images["v1.13.12"]="kindest/node:v1.13.12@sha256:1fe072c080ee129a2a440956a65925ab3bbd1227cf154e2ade145b8e59a584ad"
kind_node_images["v1.14.9"]="kindest/node:v1.14.9@sha256:bdd3731588fa3ce8f66c7c22f25351362428964b6bca13048659f68b9e665b72"
kind_node_images["v1.15.6"]="kindest/node:v1.15.6@sha256:18c4ab6b61c991c249d29df778e651f443ac4bcd4e6bdd37e0c83c0d33eaae78"
kind_node_images["v1.16.3"]="kindest/node:v1.16.3@sha256:70ce6ce09bee5c34ab14aec2b84d6edb260473a60638b1b095470a3a0f95ebec"
kind_node_images["v1.17.0"]="kindest/node:v1.17.0@sha256:190c97963ec4f4121c3f1e96ca6eb104becda5bae1df3a13f01649b2dd372f6d"

function e2e::image_build() {
    if [ -n "$SKIP_BUILD" ]; then
        echo "info: skip building images"
        export NO_BUILD=y
    fi
    if [ -n "$SKIP_IMAGE_BUILD" ]; then
        echo "info: skip building and pushing images"
        return
    fi
    DOCKER_REGISTRY=$DOCKER_REGISTRY GOOS=linux GOARCH=amd64 make e2e-docker
}

function e2e::image_load() {
    local names=(
        pingcap/chaos-mesh-e2e
    )
    for n in ${names[@]}; do
        $KIND_BIN load docker-image --name $CLUSTER $DOCKER_REGISTRY/$n:$IMAGE_TAG
    done
}

function e2e::cluster_exists() {
    local name="$1"
    $KIND_BIN get clusters | grep $CLUSTER &>/dev/null
}

function e2e::__restart_docker() {
    echo "info: restarting docker"
    service docker restart
    # the service can be started but the docker socket not ready, wait for ready
    local WAIT_N=0
    local MAX_WAIT=5
    while true; do
        # docker ps -q should only work if the daemon is ready
        docker ps -q > /dev/null 2>&1 && break
        if [[ ${WAIT_N} -lt ${MAX_WAIT} ]]; then
            WAIT_N=$((WAIT_N+1))
            echo "info; Waiting for docker to be ready, sleeping for ${WAIT_N} seconds."
            sleep ${WAIT_N}
        else
            echo "info: Reached maximum attempts, not waiting any longer..."
            break
        fi
    done
    echo "info: done restarting docker"
}

# e2e::__cluster_is_alive checks if the cluster is alive or not
function e2e::__cluster_is_alive() {
    local ret=0
    echo "info: checking the cluster version"
    $KUBECTL_BIN --context $KUBECONTEXT version --short || ret=$?
    return $ret
}

function e2e::up() {
    if [ -n "$SKIP_UP" ]; then
        echo "info: skip starting a new cluster"
        return
    fi
    if [ -n "$DOCKER_IO_MIRROR" -a -n "${DOCKER_IN_DOCKER_ENABLED:-}" ]; then
        echo "info: configure docker.io mirror '$DOCKER_IO_MIRROR' for DinD"
cat <<EOF > /etc/docker/daemon.json
{
    "registry-mirrors": ["$DOCKER_IO_MIRROR"]
}
EOF
        e2e::__restart_docker
    fi
    if e2e::cluster_exists $CLUSTER; then
        if [ -n "$REUSE_CLUSTER" ]; then
            if e2e::__cluster_is_alive; then
                echo "info: REUSE_CLUSTER is enabled and the cluster is alive, reusing it"
                return
            else
                echo "info: REUSE_CLUSTER is enabled but the cluster is not alive, trying to recreate it"
            fi
        fi
        echo "info: deleting the cluster '$CLUSTER'"
        $KIND_BIN delete cluster --name $CLUSTER
    fi
    echo "info: starting a new cluster"
    tmpfile=$(mktemp)
    trap "test -f $tmpfile && rm $tmpfile" RETURN
    cat <<EOF > $tmpfile
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
EOF
    if [ -n "$DOCKER_IO_MIRROR" -o -n "$GCR_IO_MIRROR" -o -n "$QUAY_IO_MIRROR" ]; then
cat <<EOF >> $tmpfile
containerdConfigPatches:
- |-
EOF
        if [ -n "$DOCKER_IO_MIRROR" ]; then
cat <<EOF >> $tmpfile
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
    endpoint = ["$DOCKER_IO_MIRROR"]
EOF
        fi
        if [ -n "$GCR_IO_MIRROR" ]; then
cat <<EOF >> $tmpfile
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."gcr.io"]
    endpoint = ["$GCR_IO_MIRROR"]
EOF
        fi
        if [ -n "$QUAY_IO_MIRROR" ]; then
cat <<EOF >> $tmpfile
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."quay.io"]
    endpoint = ["$QUAY_IO_MIRROR"]
EOF
        fi
    fi
    # control-plane
    cat <<EOF >> $tmpfile
nodes:
- role: control-plane
EOF
    if [[ "$KIND_DATA_HOSTPATH" != "none" ]]; then
        if [ ! -d "$KIND_DATA_HOSTPATH" ]; then
            echo "error: '$KIND_DATA_HOSTPATH' is not a directory"
            exit 1
        fi
        local hostWorkerPath="${KIND_DATA_HOSTPATH}/control-plane"
        test -d $hostWorkerPath || mkdir $hostWorkerPath
        cat <<EOF >> $tmpfile
  extraMounts:
  - containerPath: /mnt/disks/
    hostPath: "$hostWorkerPath"
    propagation: HostToContainer
EOF
    fi
    # workers
    for ((i = 1; i <= $KUBE_WORKERS; i++)) {
        cat <<EOF >> $tmpfile
- role: worker
EOF
        if [[ "$KIND_DATA_HOSTPATH" != "none" ]]; then
            if [ ! -d "$KIND_DATA_HOSTPATH" ]; then
                echo "error: '$KIND_DATA_HOSTPATH' is not a directory"
                exit 1
            fi
            local hostWorkerPath="${KIND_DATA_HOSTPATH}/worker${i}"
            test -d $hostWorkerPath || mkdir $hostWorkerPath
            cat <<EOF >> $tmpfile
  extraMounts:
  - containerPath: /mnt/disks/
    hostPath: "$hostWorkerPath"
    propagation: HostToContainer
EOF
        fi
    }
    echo "info: print the contents of kindconfig"
    cat $tmpfile
    echo "info: end of the contents of kindconfig"
    echo "info: creating the cluster '$CLUSTER'"
    local image=""
    for v in ${!kind_node_images[*]}; do
        if [[ "$KUBE_VERSION" == "$v" ]]; then
            image=${kind_node_images[$v]}
            echo "info: image for $KUBE_VERSION: $image"
            break
        fi
    done
    if [ -z "$image" ]; then
        echo "error: no image for $KUBE_VERSION, exit"
        exit 1
    fi
    $KIND_BIN create cluster --config $KUBECONFIG --name $CLUSTER --image $image --config $tmpfile -v 4
    # make it able to schedule pods on control-plane, then less resources we required
    # This is disabled because when hostNetwork is used, pd requires 2379/2780
    # which may conflict with etcd on control-plane.
    #echo "info: remove 'node-role.kubernetes.io/master' taint from $CLUSTER-control-plane"
    #kubectl taint nodes $CLUSTER-control-plane node-role.kubernetes.io/master-
}


function e2e::__wait_for_deploy() {
    local ns="$1"
    local name="$2"
    local retries="${3:-300}"
    echo "info: waiting for pods of deployment $ns/$name are ready (retries: $retries, interval: 1s)"
    for ((i = 0; i < retries; i++)) {
        read a b <<<$($KUBECTL_BIN --context $KUBECONTEXT -n $ns get deploy/$name -ojsonpath='{.spec.replicas} {.status.readyReplicas}{"\n"}')
        if [[ "$a" -gt 0 && "$a" -eq "$b" ]]; then
            echo "info: all pods of deployment $ns/$name are ready (desired: $a, ready: $b)"
            return 0
        fi
        echo "info: pods of deployment $ns/$name (desired: $a, ready: $b)"
        sleep 1
    }
    echo "info: timed out waiting for pods of deployment $ns/$name are ready"
    return 1
}

function e2e::setup_helm_server() {
    $KUBECTL_BIN --context $KUBECONTEXT apply -f ${ROOT}/manifests/tiller-rbac.yaml
    if hack::version_ge $KUBE_VERSION "v1.16.0"; then
        # workaround for https://github.com/helm/helm/issues/6374
        # TODO remove this when we can upgrade to helm 2.15+, see https://github.com/helm/helm/pull/6462
        $HELM_BIN init --service-account tiller --output yaml \
            | sed 's@apiVersion: extensions/v1beta1@apiVersion: apps/v1@' \
            | sed 's@  replicas: 1@  replicas: 1\n  selector: {"matchLabels": {"app": "helm", "name": "tiller"}}@' \
            | $KUBECTL_BIN --context $KUBECONTEXT apply -f -
        echo "info: wait for tiller to be ready"
        e2e::__wait_for_deploy kube-system tiller-deploy
    else
        $HELM_BIN init --service-account=tiller --wait
    fi
    $HELM_BIN version
}

function e2e::down() {
    if [ -n "$SKIP_DOWN" ]; then
        echo "info: skip shutting down the cluster '$CLUSTER'"
        return
    fi
    if ! e2e::cluster_exists $CLUSTER; then
        echo "info: cluster '$CLUSTER' does not exist, skip shutting down the cluster"
        return
    fi
    $KIND_BIN delete cluster --name $CLUSTER
}

trap "e2e::down" EXIT

e2e::up
e2e::setup_helm_server
e2e::image_build
e2e::image_load

export KUBECONFIG
export KUBECONTEXT
export E2E_IMAGE=$DOCKER_REGISTRY/pingcap/chaos-mesh-e2e:${IMAGE_TAG}

hack/run-e2e.sh "$@"