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

set -eu

cur=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd $cur

kubectl apply -f ./rbac.yaml

SA_SECRET_NAME=$(kubectl get serviceaccounts fake-sa -ojsonpath='{.secrets[0].name}')
SA_SECRET=$(kubectl get secrets "${SA_SECRET_NAME}" -o=jsonpath='{.data.token}' | base64 -d)

kubectl config set-credentials fake-sa --token "${SA_SECRET}"

CURRENT_CONTEXT=$(kubectl config current-context)
# line 2, column 3
CURRENT_CLUSTER=$(kubectl config get-contexts "${CURRENT_CONTEXT}" |awk 'NR==2'| awk '{print $3}')

kubectl config set-context fake-sa-test --cluster "${CURRENT_CLUSTER}" --user fake-sa

kubectl --context fake-sa-test auth can-i create podchaos || exit 1
kubectl --context fake-sa-test auth can-i get podchaos && exit 1

kubectl --context fake-sa-test create -f podchaos-example.yaml || exit 1
