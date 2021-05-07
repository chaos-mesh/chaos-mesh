#!/usr/bin/env bash

# Copyright 2020 Chaos Mesh Authors.
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
    PROVIDER                    Kubernetes provider, e.g. kind, gke, eks, defaults: kind
    DOCKER_REGISTRY             image docker registry
    IMAGE_TAG                   image tag
    SKIP_BUILD                  skip building binaries
    SKIP_IMAGE_BUILD            skip build and push images
    SKIP_UP                     skip starting the cluster
    SKIP_DOWN                   skip shutting down the cluster
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
    SKIP_UP=y SKIP_DOWN=y ./hack/e2e.sh -- <e2e args>
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
echo "ensured kind"
hack::ensure_kubectl
echo "ensured kubectl"
hack::ensure_kubebuilder
echo "ensured kubebuilder"
hack::ensure_kustomize
echo "ensured kustomize"
hack::ensure_kubetest2
echo "ensured kubetest2"

PROVIDER=${PROVIDER:-kind}
DOCKER_REGISTRY=${DOCKER_REGISTRY:-localhost:5000}
IMAGE_TAG=${IMAGE_TAG:-latest}
CLUSTER=${CLUSTER:-chaos-mesh}
KUBECONFIG=${KUBECONFIG:-~/.kube/config}
SKIP_BUILD=${SKIP_BUILD:-}
SKIP_IMAGE_BUILD=${SKIP_IMAGE_BUILD:-}
SKIP_UP=${SKIP_UP:-}
SKIP_DOWN=${SKIP_DOWN:-}
SKIP_DUMP=${SKIP_DUMP:-}
SKIP_TEST=${SKIP_TEST:-}
KIND_DATA_HOSTPATH=${KIND_DATA_HOSTPATH:-none}
KUBE_VERSION=${KUBE_VERSION:-v1.20.2}
KUBE_WORKERS=${KUBE_WORKERS:-3}
DOCKER_IO_MIRROR=${DOCKER_IO_MIRROR:-}
GCR_IO_MIRROR=${GCR_IO_MIRROR:-}
QUAY_IO_MIRROR=${QUAY_IO_MIRROR:-}
RUNNER_SUITE_NAME=${RUNNER_SUITE_NAME:-}
ARTIFACTS=${ARTIFACTS:-}

echo "PROVIDER: $PROVIDER"
echo "DOCKER_REGISTRY: $DOCKER_REGISTRY"
echo "IMAGE_TAG: $IMAGE_TAG"
echo "CLUSTER: $CLUSTER"
echo "KUBECONFIG: $KUBECONFIG"
echo "SKIP_BUILD: $SKIP_BUILD"
echo "SKIP_IMAGE_BUILD: $SKIP_IMAGE_BUILD"
echo "SKIP_UP: $SKIP_UP"
echo "SKIP_DOWN: $SKIP_DOWN"
echo "SKIP_DUMP: $SKIP_DUMP"
echo "KIND_DATA_HOSTPATH: $KIND_DATA_HOSTPATH"
echo "KUBE_VERSION: $KUBE_VERSION"
echo "DOCKER_IO_MIRROR: $DOCKER_IO_MIRROR"
echo "GCR_IO_MIRROR: $GCR_IO_MIRROR"
echo "QUAY_IO_MIRROR: $QUAY_IO_MIRROR"
echo "ARTIFACTS: $ARTIFACTS"
echo "KUBE_WORKERS: $KUBE_WORKERS"

# https://github.com/kubernetes-sigs/kind/releases/tag/v0.8.1
declare -A kind_node_images
kind_node_images["v1.11.10"]="kindest/node:v1.11.10@sha256:74c8740710649a3abb169e7f348312deff88fc97d74cfb874c5095ab3866bb42"
kind_node_images["v1.12.10"]="kindest/node:v1.12.10@sha256:faeb82453af2f9373447bb63f50bae02b8020968e0889c7fa308e19b348916cb"
kind_node_images["v1.13.12"]="kindest/node:v1.13.12@sha256:214476f1514e47fe3f6f54d0f9e24cfb1e4cda449529791286c7161b7f9c08e7"
kind_node_images["v1.14.10"]="kindest/node:v1.14.10@sha256:6cd43ff41ae9f02bb46c8f455d5323819aec858b99534a290517ebc181b443c6"
kind_node_images["v1.15.11"]="kindest/node:v1.15.11@sha256:6cc31f3533deb138792db2c7d1ffc36f7456a06f1db5556ad3b6927641016f50"
kind_node_images["v1.16.9"]="kindest/node:v1.16.9@sha256:7175872357bc85847ec4b1aba46ed1d12fa054c83ac7a8a11f5c268957fd5765"
kind_node_images["v1.17.5"]="kindest/node:v1.17.5@sha256:ab3f9e6ec5ad8840eeb1f76c89bb7948c77bbf76bcebe1a8b59790b8ae9a283a"
kind_node_images["v1.18.2"]="kindest/node:v1.18.2@sha256:7b27a6d0f2517ff88ba444025beae41491b016bc6af573ba467b70c5e8e0d85f"
kind_node_images["v1.20.2"]="kindest/node:v1.20.2@sha256:8f7ea6e7642c0da54f04a7ee10431549c0257315b3a634f6ef2fecaaedb19bab"

function e2e::image_build() {
    if [ -n "$SKIP_BUILD" ]; then
        echo "info: skip building images"
        export NO_BUILD=y
    fi
    if [ -n "$SKIP_IMAGE_BUILD" ]; then
        echo "info: skip building and pushing images"
        return
    fi
    DOCKER_REGISTRY=${DOCKER_REGISTRY} GOOS=linux GOARCH=amd64 make image-chaos-mesh-e2e
    DOCKER_REGISTRY=${DOCKER_REGISTRY} make image-chaos-mesh
    DOCKER_REGISTRY=${DOCKER_REGISTRY} make image-chaos-daemon
    DOCKER_REGISTRY=${DOCKER_REGISTRY} make image-e2e-helper
}

function e2e::create_kindconfig() {
    local tmpfile=${1}
    cat <<EOF > $tmpfile
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
kubeadmConfigPatches:
- |
  kind: ClusterConfiguration
  apiVersion: kubeadm.k8s.io/v1beta1
  apiServer:
    extraArgs:
      v: "4"
  scheduler:
    extraArgs:
      v: "4"
  controllerManager:
    extraArgs:
      v: "4"
- |
  kind: ClusterConfiguration
  apiVersion: kubeadm.k8s.io/v1beta2
  apiServer:
    extraArgs:
      v: "4"
  scheduler:
    extraArgs:
      v: "4"
  controllerManager:
    extraArgs:
      v: "4"
- |
  # backward compatibility for Kubernetes 1.12 and prior versions
  kind: ClusterConfiguration
  apiVersion: kubeadm.k8s.io/v1alpha3
  apiServerExtraArgs:
    v: "4"
  schedulerExtraArgs:
    v: "4"
  controllerManagerExtraArgs:
    v: "4"
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
}

e2e::image_build

kubetest2_args=(
    $PROVIDER
)

if [ -n "$RUNNER_SUITE_NAME" ]; then
    kubetest2_args+=(
        --suite-name "$RUNNER_SUITE_NAME"
    )
fi

if [ -z "$SKIP_UP" ]; then
    kubetest2_args+=(--up)
fi

if [ -z "$SKIP_DOWN" ]; then
    kubetest2_args+=(--down)
fi

if [ -z "$SKIP_TEST" ]; then
    kubetest2_args+=(--test exec)
fi

if [ "$PROVIDER" == "kind" ]; then
    tmpfile=$(mktemp)
    trap "test -f $tmpfile && rm $tmpfile" EXIT
    e2e::create_kindconfig $tmpfile
    echo "info: print the contents of kindconfig"
    cat $tmpfile
    image=""
    for v in ${!kind_node_images[*]}; do
        if [[ "$KUBE_VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ && "$KUBE_VERSION" == "$v" ]]; then
            image=${kind_node_images[$v]}
            echo "info: image for $KUBE_VERSION: $image"
        elif [[ "$KUBE_VERSION" =~ ^v[0-9]+\.[0-9]+$ && "$KUBE_VERSION" == "${v%.*}" ]]; then
            image=${kind_node_images[$v]}
            echo "info: image for $KUBE_VERSION: $image"
        fi
    done
    if [ -z "$image" ]; then
        echo "error: no image for $KUBE_VERSION, exit"
        exit 1
    fi
    kubetest2_args+=(--image-name $image)
    kubetest2_args+=(
        # add some retires because kind may fail to start the cluster when the
        # load is high
        --up-retries 3
        --cluster-name "$CLUSTER"
        --config "$tmpfile"
        --verbosity 4
    )
fi

export PROVIDER
export CLUSTER
export KUBECONFIG
export E2E_IMAGE=${DOCKER_REGISTRY}/pingcap/chaos-mesh-e2e:${IMAGE_TAG}
export DOCKER_REGISTRY=${DOCKER_REGISTRY}
export IMAGE_TAG=${IMAGE_TAG}
export PATH=$OUTPUT_BIN:$PATH

if [ -n "${ARTIFACTS}" ]; then
    export REPORT_DIR=${ARTIFACTS}
fi

if [ -n "${ARTIFACTS}" -a -z "$SKIP_DUMP" ]; then
    kubetest2_args+=(--dump)
fi

echo "info: run 'kubetest2 ${kubetest2_args[@]} -- hack/run-e2e.sh $@'"
$KUBETSTS2_BIN ${kubetest2_args[@]} -- hack/run-e2e.sh "$@"
