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

set -e

ROOT=$(unset CDPATH && cd $(dirname "${BASH_SOURCE[0]}")/.. && pwd)
cd $ROOT

usage() {
    cat <<EOF
This script use kind to create Kubernetes cluster,about kind please refer: https://kind.sigs.k8s.io/
Before run this script,please ensure that:
* have installed docker
* have installed helm
Options:
       -h,--help               prints the usage message
       -n,--name               name of the Kubernetes cluster, default value: kind
       -c,--nodeNum            the count of the cluster nodes, default value: 3
       -k,--k8sVersion         version of the Kubernetes cluster, default value: v1.20.7
       -v,--volumeNum          the volumes number of each kubernetes node, default value: 5
       -r,--registryName       the name of local docker registry, default value: registry
       -p,--registryPort       the published port of local docker registry, default value: 5000
Usage:
    $0 --name testCluster --nodeNum 4 --k8sVersion v1.20.7
EOF
}

while [[ $# -gt 0 ]]
do
key="$1"

case $key in
    -n|--name)
    clusterName="$2"
    shift
    shift
    ;;
    -c|--nodeNum)
    nodeNum="$2"
    shift
    shift
    ;;
    -k|--k8sVersion)
    k8sVersion="$2"
    shift
    shift
    ;;
    -v|--volumeNum)
    volumeNum="$2"
    shift
    shift
    ;;
    -r|--registryName)
    registryName="$2"
    shift
    shift
    ;;
    -p|--registryPort)
    registryPort="$2"
    shift
    shift
    ;;
    -h|--help)
    usage
    exit 0
    ;;
    *)
    echo "unknown option: $key"
    usage
    exit 1
    ;;
esac
done

clusterName=${clusterName:-kind}
nodeNum=${nodeNum:-3}
k8sVersion=${k8sVersion:-v1.20.7}
volumeNum=${volumeNum:-5}
registryName=${registryName:-registry}
registryPort=${registryPort:-5000}

echo "clusterName: ${clusterName}"
echo "nodeNum: ${nodeNum}"
echo "k8sVersion: ${k8sVersion}"
echo "volumeNum: ${volumeNum}"
echo "registryName: ${registryName}"
echo "registryPort: ${registryPort}"

source "${ROOT}/hack/lib.sh"

echo "ensuring kind"
hack::ensure_kind
echo "ensuring kubectl"
hack::ensure_kubectl

OUTPUT_BIN=${ROOT}/output/bin
KUBECTL_BIN=${OUTPUT_BIN}/kubectl
HELM_BIN=${OUTPUT_BIN}/helm
KIND_BIN=${OUTPUT_BIN}/kind

# create registry container unless it already exists
running="$(docker inspect -f '{{.State.Running}}' "${registryName}" 2>/dev/null || true)"
if [ "${running}" != 'true' ]; then
  docker run \
    -d --restart=always -p "${registryPort}:5000" --name "${registryName}" \
    registry:2
fi

echo "############# start create cluster:[${clusterName}] #############"
workDir=${HOME}/kind/${clusterName}
kubeconfigPath=${workDir}/config
mkdir -p ${workDir}

data_dir=${workDir}/data

echo "clean data dir: ${data_dir}"
if [ -d ${data_dir} ]; then
    rm -rf ${data_dir}
fi

configFile=${workDir}/kind-config.yaml

cat <<EOF > ${configFile}
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
kubeadmConfigPatches:
- |
  apiVersion: kubeadm.k8s.io/v1beta2
  kind: ClusterConfiguration
  metadata:
    name: config
  apiServerExtraArgs:
    enable-admission-plugins: NodeRestriction,MutatingAdmissionWebhook,ValidatingAdmissionWebhook
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:${registryPort}"]
    endpoint = ["http://${registryName}:${registryPort}"]
nodes:
- role: control-plane
EOF

for ((i=0;i<nodeNum;i++))
do
    mkdir -p ${data_dir}/worker${i}
    cat <<EOF >>  ${configFile}
- role: worker
  extraMounts:
EOF
    for ((k=1;k<=volumeNum;k++))
    do
        mkdir -p ${data_dir}/worker${i}/vol${k}
        cat <<EOF >> ${configFile}
  - containerPath: /mnt/disks/vol${k}
    hostPath: ${data_dir}/worker${i}/vol${k}
EOF
    done
done

echo "start to create k8s cluster"
${KIND_BIN} create cluster --config ${configFile} --image kindest/node:${k8sVersion} --name=${clusterName}
${KIND_BIN} get kubeconfig --name=${clusterName} > ${kubeconfigPath}
export KUBECONFIG=${kubeconfigPath}

echo "connect the local docker registry to the cluster network"

set +e

connected=$(docker network connect "kind" "${registryName}" 2>&1)
exitCode=$?

if [[ exitCode -ne 0 ]] && [[ $connected != *"already exists"* ]]; then
  echo "error when connecting docker registry: ${connected}"
  exit 1
fi

set -e

${KUBECTL_BIN} apply -f ${ROOT}/manifests/local-volume-provisioner.yaml

$KUBECTL_BIN create ns chaos-mesh

echo "############# success create cluster:[${clusterName}] #############"

echo "To start using your cluster, run:"
echo "    export KUBECONFIG=${kubeconfigPath}"
echo ""
cat << EOF
NOTE: In kind, nodes run docker network and cannot access host network.
If you configured local HTTP proxy in your docker, images may cannot be pulled
because http proxy is inaccessible.
If you cannot remove http proxy settings, you can either whitelist image
domains in NO_PROXY environment or use 'docker pull <image> && kind load
docker-image <image>' command to load images into nodes.
EOF
