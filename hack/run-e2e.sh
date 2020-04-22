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

set -o errexit
set -o nounset
set -o pipefail

ROOT=$(unset CDPATH && cd $(dirname "${BASH_SOURCE[0]}")/.. && pwd)
cd $ROOT

source $ROOT/hack/lib.sh

hack::ensure_kubectl
hack::ensure_helm

E2E_IMAGE=${E2E_IMAGE:-localhost:5000/pingcap/chaos-mesh-e2e:latest}
KUBECONFIG=${KUBECONFIG:-$HOME/.kube/config}
KUBECONTEXT=${KUBECONTEXT:-}
REPORT_DIR=${REPORT_DIR:-}
REPORT_PREFIX=${REPORT_PREFIX:-}
DOCKER_REGISTRY=${DOCKER_REGISTRY:-localhost:5000}
IMAGE_TAG=${IMAGE_TAG:-latest}

if [ -z "$KUBECONFIG" ]; then
    echo "error: KUBECONFIG is required"
    exit 1
fi

echo "E2E_IMAGE: $E2E_IMAGE"
echo "KUBECONFIG: $KUBECONFIG"
echo "KUBECONTEXT: $KUBECONTEXT"
echo "REPORT_DIR: $REPORT_DIR"
echo "REPORT_PREFIX: $REPORT_PREFIX"
echo "DOCKER_REGISTRY: $DOCKER_REGISTRY"
echo "IMAGE_TAG: $IMAGE_TAG"



GINKGO_PARALLEL=${GINKGO_PARALLEL:-n} # set to 'y' to run tests in parallel
# If 'y', Ginkgo's reporter will not print out in color when tests are run
# in parallel
GINKGO_NO_COLOR=${GINKGO_NO_COLOR:-n}
GINKGO_STREAM=${GINKGO_STREAM:-y}

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

echo "info: start to run e2e process"
e2e_args=(
    /usr/local/bin/ginkgo
    ${ginkgo_args[@]:-}
    /usr/local/bin/e2e.test
    --
    --manager-image="${DOCKER_REGISTRY}/pingcap/chaos-mesh"
    --manager-image-tag="${IMAGE_TAG}"
    --daemon-image="${DOCKER_REGISTRY}/pingcap/chaos-daemon"
    --daemon-image-tag="${IMAGE_TAG}"
    --timer-image="${DOCKER_REGISTRY}/pingcap/timer:latest"
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
    docker_args+=(
        -v $REPORT_DIR:$REPORT_DIR
    )
fi

echo "info: docker ${docker_args[@]} $E2E_IMAGE ${e2e_args[@]}"
docker ${docker_args[@]} $E2E_IMAGE ${e2e_args[@]}
