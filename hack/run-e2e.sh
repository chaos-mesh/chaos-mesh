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

set -o errexit
set -o nounset
set -o pipefail

ROOT=$(unset CDPATH && cd $(dirname "${BASH_SOURCE[0]}")/.. && pwd)
cd $ROOT

source $ROOT/hack/lib.sh

hack::ensure_kubectl
hack::ensure_helm
hack::ensure_kind


PROVIDER=${PROVIDER:-}
CLUSTER=${CLUSTER:-}
IMAGE_TAG=${IMAGE_TAG:-}
E2E_IMAGE=${E2E_IMAGE:-localhost:5000/pingcap/chaos-mesh-e2e:latest}
KUBECONFIG=${KUBECONFIG:-$HOME/.kube/config}
KUBECONTEXT=${KUBECONTEXT:-kind-chaos-mesh}
REPORT_DIR=${REPORT_DIR:-}
REPORT_PREFIX=${REPORT_PREFIX:-}
GINKGO_NODES=${GINKGO_NODES:-}
GINKGO_PARALLEL=${GINKGO_PARALLEL:-n} # set to 'y' to run tests in parallel
# If 'y', Ginkgo's reporter will not print out in color when tests are run
# in parallel
GINKGO_NO_COLOR=${GINKGO_NO_COLOR:-n}
GINKGO_STREAM=${GINKGO_STREAM:-y}
SKIP_GINKGO=${SKIP_GINKGO:-}
# We don't finalizers namespace on failure by default for easier debugging in local development.
# TODO support this feature
DELETE_NAMESPACE_ON_FAILURE=${DELETE_NAMESPACE_ON_FAILURE:-false}
DOCKER_REGISTRY=${DOCKER_REGISTRY:-localhost:5000}

if [ -z "$KUBECONFIG" ]; then
    echo "error: KUBECONFIG is required"
    exit 1
fi

echo "PROVIDER: $PROVIDER"
echo "CLUSTER: $CLUSTER"
echo "IMAGE_TAG: $IMAGE_TAG"
echo "E2E_IMAGE: $E2E_IMAGE"
echo "KUBECONFIG: $KUBECONFIG"
echo "KUBECONTEXT: $KUBECONTEXT"
echo "REPORT_DIR: $REPORT_DIR"
echo "REPORT_PREFIX: $REPORT_PREFIX"
echo "DELETE_NAMESPACE_ON_FAILURE: $DELETE_NAMESPACE_ON_FAILURE"
echo "DOCKER_REGISTRY: $DOCKER_REGISTRY"

function e2e::image_load() {
    local images=(
        pingcap/chaos-mesh
        pingcap/chaos-daemon
        pingcap/e2e-helper
    )
    if [ "$PROVIDER" == "kind" ]; then
        local nodes=$($KIND_BIN get nodes --name $CLUSTER | grep -v 'control-plane$')
        echo $nodes
        echo "info: load images ${images[@]}"
        for image in ${images[@]}; do
            $KIND_BIN load docker-image --name $CLUSTER ${DOCKER_REGISTRY}/$image:$IMAGE_TAG --nodes $(hack::join ',' ${nodes[@]})
        done

        # bypassing docker pull rate limit inner the kind container: kindest/node has no credentials
        # pingcap/coredns:latest, nginx:latest and gcr.io/google-containers/pause:latest is required for test
        # we suppose that you could pull this image on your host docker
        echo "info: load images pingcap/coredns:latest, nginx:latest and gcr.io/google-containers/pause:latest"
        docker pull pingcap/coredns:v0.2.0
        docker pull nginx:latest
        docker pull gcr.io/google-containers/pause:latest
        $KIND_BIN load docker-image --name $CLUSTER pingcap/coredns:v0.2.0 --nodes $(hack::join ',' ${nodes[@]})
        $KIND_BIN load docker-image --name $CLUSTER nginx:latest --nodes $(hack::join ',' ${nodes[@]})
        $KIND_BIN load docker-image --name $CLUSTER gcr.io/google-containers/pause:latest --nodes $(hack::join ',' ${nodes[@]})
    fi
}


function e2e::get_kube_version() {
    $KUBECTL_BIN --context $KUBECONTEXT version --short | awk '/Server Version:/ {print $3}'
}

if [ -z "$KUBECONTEXT" ]; then
    echo "info: KUBECONTEXT is not set, current context is used"
    KUBECONTEXT=$($KUBECTL_BIN config current-context 2>/dev/null) || true
    if [ -z "$KUBECONTEXT" ]; then
        echo "error: current context cannot be detected"
        exit 1
    fi
    echo "info: current kubeconfig context is '$KUBECONTEXT'"
fi

e2e::image_load
echo "info: image loaded"

if [ -n "$SKIP_GINKGO" ]; then
    echo "info: skipping ginkgo"
    exit 0
fi

echo "info: start to run e2e process"

ginkgo_args=()

if [[ -n "${GINKGO_NODES:-}" ]]; then
    ginkgo_args+=("--nodes=${GINKGO_NODES}")
elif [[ ${GINKGO_PARALLEL} =~ ^[yY]$ ]]; then
    ginkgo_args+=("-p")
fi

if [[ "${GINKGO_NO_COLOR}" == "y" ]]; then
    ginkgo_args+=("--noColor")
fi

if [[ "${GINKGO_STREAM}" == "y" ]]; then
    ginkgo_args+=("--stream")
fi

e2e_args=(
    /usr/local/bin/ginkgo
    ${ginkgo_args[@]:-}
    /usr/local/bin/e2e.test
    --
    --manager-image="${DOCKER_REGISTRY}/pingcap/chaos-mesh"
    --manager-image-tag="${IMAGE_TAG}"
    --daemon-image="${DOCKER_REGISTRY}/pingcap/chaos-daemon"
    --daemon-image-tag="${IMAGE_TAG}"
    --e2e-image="${DOCKER_REGISTRY}/pingcap/e2e-helper:${IMAGE_TAG}"
    --install-chaos-mesh
)

if [ -n "$REPORT_DIR" ]; then
    e2e_args+=(
        --report-dir="${REPORT_DIR}"
        --report-prefix="${REPORT_PREFIX}"
    )
fi

e2e_args+=(${@:-})

docker_args=(
    run
    --rm
    --net=host
    --privileged
    -v /:/rootfs
    -v $ROOT:$ROOT
    -w $ROOT
    -v $KUBECONFIG:/etc/kubernetes/admin.conf:ro
    --env KUBECONFIG=/etc/kubernetes/admin.conf
    --env KUBECONTEXT=$KUBECONTEXT
)

if [ -n "$REPORT_DIR" ]; then
    e2e_args+=(
        --report-dir="${REPORT_DIR}"
        --report-prefix="${REPORT_PREFIX}"
    )
    docker_args+=(
        -v $REPORT_DIR:$REPORT_DIR
    )
fi

echo "info: docker ${docker_args[@]} $E2E_IMAGE ${e2e_args[@]}"
docker ${docker_args[@]} $E2E_IMAGE ${e2e_args[@]}
