#!/usr/bin/env bash

# Copyright 2021 Chaos Mesh Authors.
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
set -e

NAMESPACES=$(kubectl get namespace | sed '1d' | awk '{print $1}')
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
CRDS="awschaos
dnschaos
gcpchaos
iochaos
jvmchaos
kernelchaos
networkchaos
podchaos
stresschaos
timechaos
"
cnt=0

usage() {
    cat << EOF
This script is used to migrate to chaos-mesh 2.0.
USAGE:
    migrate.sh [FLAGS]
FLAGS:
    -h, --help              Prints help information
    -e, --export            Export existing chaos and update them
    -i, --import            Import updated chaos (do this after chaos-mesh upgrade)
    -c, --crd               Update CRD (do this after exporting chaos, and before ugrading chaos-mesh)
EOF
}

build () {
    cd $SCRIPT_DIR
    go build main.go
}

update_yaml () {
    local yaml=$1
    ./main $yaml $yaml
}

reapply_crd () {
    local crd=""
    kubectl delete -f https://mirrors.chaos-mesh.org/v1.2.1/crd.yaml
    if kubectl api-versions | grep -q -w apiextensions.k8s.io/v1 ; then
        crd="https://mirrors.chaos-mesh.org/latest/crd.yaml"
    else
        crd="https://mirrors.chaos-mesh.org/latest/crd-v1beta1.yaml"
    fi
    kubectl create -f ${crd}
}

handle_namespace () {
    local namespace=$1
    for kind in $CRDS
    do
        echo "  searching resources $kind"
        resources=$(kubectl get $kind -n $namespace | sed '1d' | awk '{print $1}')
        for resource in $resources
        do
            echo "      getting $resource"
            kubectl get $kind $resource -n $namespace -o yaml > $cnt.yaml
            update_yaml $cnt.yaml
            let cnt++
        done
    done
}

export_chaos () {
    build

    for ns in $NAMESPACES
    do
        echo "searching namespace $ns"
        handle_namespace $ns
    done
}

import_chaos() {
    local yamls=$(find . -regex ".*\.yaml")
    for yaml in $yamls
    do
        kubectl apply -f $yaml
    done
}

UPDATE_CRD=false
EXPORT_CHAOS=false
IMPORT_CHAOS=false

while [[ $# -gt 0 ]]
do
key="$1"

case $key in
    -e|--export)
        EXPORT_CHAOS=true
        shift
        ;;
    -i|--import)
        IMPORT_CHAOS=true
        shift
        ;;
    -c|--crd)
        UPDATE_CRD=true
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

if [ ${EXPORT_CHAOS} == "true" ]; then
    export_chaos
fi

if [ ${UPDATE_CRD} == "true" ]; then
    reapply_crd
fi

if [ ${IMPORT_CHAOS} == "true" ]; then
    import_chaos
fi
