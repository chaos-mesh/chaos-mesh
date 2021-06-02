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

set -u
code=0
cur=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd $cur/../../../bin

pwd
echo "deploy deployments and chaos for testing"
wget https://mirrors.chaos-mesh.org/v1.1.2/web-show/deploy.sh
bash deploy.sh
cat <<EOF >delay.yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: web-show-network-delay
spec:
  action: delay # the specific chaos action to inject
  mode: one # the mode to run chaos action; supported modes are one/all/fixed/fixed-percent/random-max-percent
  selector: # pods where to inject chaos actions
    namespaces:
      - default
    labelSelectors:
      "app": "web-show"  # the label of the pod for chaos injection
  delay:
    latency: "10ms"
  duration: "30s" # duration for the injected chaos experiment
EOF
kubectl apply -f delay.yaml

echo "Checking chaosctl function"
./chaosctl logs 1>/dev/null
status=$(./chaosctl debug -i networkchaos web-show-network-delay | grep "Execute as expected")
if [[ -z "$status" ]]; then
    ./chaosctl debug -i networkchaos web-show-network-delay
    echo "Chaos is not running as expected"
    code=1
fi
echo "Cleaning up"
kubectl delete -f delay.yaml
rm delay.yaml
bash deploy.sh -d
rm deploy.sh
exit $code
