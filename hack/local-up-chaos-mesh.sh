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

set -o errexit
set -o nounset
set -o pipefail

ROOT=$(unset CDPATH && cd $(dirname "${BASH_SOURCE[0]}")/.. && pwd)
cd $ROOT

source "${ROOT}/hack/lib.sh"

function usage() {
    cat <<'EOF'
This commands run chaos-mesh in Kubernetes.

Usage: hack/local-up-operator.sh [-hd]

    -h      show this message and exit
    -i      install dependencies only

Environments:

    PROVIDER              Kubernetes provider. Defaults: kind.
    CLUSTER               the name of e2e cluster. Defaults to kind for kind provider.
    KUBECONFIG            path to the kubeconfig file, defaults: ~/.kube/config
    KUBECONTEXT           context in kubeconfig file, defaults to current context
    NAMESPACE             Kubernetes namespace in which we run our chaos-mesh.
    IMAGE_REGISTRY        image docker registry
    IMAGE_TAG             image tag
    SKIP_IMAGE_BUILD      skip build and push images

EOF
}


dependency_only=false
while getopts "h?i" opt; do
    case "$opt" in
    h|\?)
        usage
        exit 0
        ;;
    i)
      dependency_only=true
        ;;
    esac
done

PROVIDER=${PROVIDER:-kind}
CLUSTER=${CLUSTER:-}
KUBECONFIG=${KUBECONFIG:-~/.kube/config}
KUBECONTEXT=${KUBECONTEXT:-}
IMAGE_REGISTRY_PREFIX=${IMAGE_REGISTRY_PREFIX:-ghcr.io}
IMAGE_TAG=${IMAGE_TAG:-latest}
SKIP_IMAGE_BUILD=${SKIP_IMAGE_BUILD:-}
NAMESPACE=${NAMESPACE:-chaos-mesh}

hack::ensure_kubectl
hack::ensure_kind
hack::ensure_helm

if [[ "$dependency_only" == "true" ]]; then
    exit 0
fi

function hack::cluster_exists() {
    local c="$1"
    for n in $($KIND_BIN get clusters); do
        if [ "$n" == "$c" ]; then
            return 0
        fi
    done
    return 1
}

if [ "$PROVIDER" == "kind" ]; then
    if [ -z "$CLUSTER" ]; then
        CLUSTER=kind
    fi
    if ! hack::cluster_exists "$CLUSTER"; then
        echo "error: kind cluster '$CLUSTER' not found, please create it or specify the right cluster name with CLUSTER environment"
        exit 1
    fi
else
    echo "error: only kind PROVIDER is supported"
    exit 1
fi

if [ -z "$KUBECONTEXT" ]; then
    KUBECONTEXT=$(kubectl config current-context)
    echo "info: KUBECONTEXT is not set, current context $KUBECONTEXT is used"
fi

if [ -z "$SKIP_IMAGE_BUILD" ]; then
    echo "info: building docker images"
    IMAGE_REGISTRY_PREFIX=$IMAGE_REGISTRY_PREFIX IMAGE_PROJECT=chaos-mesh IMAGE_TAG=$IMAGE_TAG UI=1 SWAGGER=1 make image
else
    echo "info: skip building docker images"
fi

echo "info: loading images into cluster"
images=(
    "$IMAGE_REGISTRY_PREFIX"/chaos-mesh/chaos-mesh:"${IMAGE_TAG}"
    "$IMAGE_REGISTRY_PREFIX"/chaos-mesh/chaos-dashboard:"${IMAGE_TAG}"
    "$IMAGE_REGISTRY_PREFIX"/chaos-mesh/chaos-daemon:"${IMAGE_TAG}"
)
for n in ${images[@]}; do
    echo "info: loading image $n"
    $KIND_BIN load docker-image --name $CLUSTER $n
done

$KUBECTL_BIN -n "$NAMESPACE" delete deploy -l app.kubernetes.io/name=chaos-mesh
$KUBECTL_BIN -n "$NAMESPACE" delete pods -l app.kubernetes.io/name=chaos-mesh

${ROOT}/install.sh --runtime containerd --crd ${ROOT}/manifests/crd.yaml --version ${IMAGE_TAG} --docker-registry ${IMAGE_REGISTRY_PREFIX}
