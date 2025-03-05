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

set -u
log_file="debug.log"
code=0
cur=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

function file_must_contains() {
  substring=$1
  match=""
  if [ "$2" = "false" ]; then
      match="-v"
  fi

  grep $match "$substring" "$log_file"
  if [ "$?" != "0" ]; then
      echo "'$substring' not found in '$log_file'"
      echo "[Debug] $log_file:"
      cat $log_file
      exit 1
  fi
}

pwd
echo "Deploy deployments and chaos for testing"
wget https://mirrors.chaos-mesh.org/v1.1.2/web-show/deploy.sh
# This is a temporary fix
# TODO: remove the following line after the new deploy.sh released
sed -i "s|image: pingcap/web-show|image: ghcr.io/chaos-mesh/web-show|g" deploy.sh
bash deploy.sh

echo "Run networkchaos"

cat <<EOF >delay.yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: web-show-network-delay
spec:
  action: delay # the specific chaos action to inject
  mode: one
  selector:
    namespaces:
      - default
    labelSelectors:
      "app": "web-show"
  delay:
    latency: "10ms"
  duration: "30s"
EOF
kubectl apply -f delay.yaml

echo "Checking chaosctl logs"
./bin/chaosctl logs >$log_file 2>&1
if [ $? -ne 0 ]; then
    echo "chaosctl logs failed"
    code=1
fi
file_must_contains "Controller manager Version:" true
file_must_contains "Chaos-daemon Version:" true
file_must_contains "\[chaos-dashboard" true

echo "Checking chaosctl debug networkchaos"
./bin/chaosctl debug networkchaos web-show-network-delay >$log_file 2>&1
if [ $? -ne 0 ]; then
    echo "chaosctl debug networkchaos failed, exit code: $?, log file: $log_file"
    echo "log file content:"
    cat $log_file
    code=1
fi
file_must_contains "\[Chaos\]: web-show-network-delay" true
file_must_contains "1. \[ipset list\]" true
file_must_contains "2. \[tc qdisc list\]" true
file_must_contains "3. \[iptables list\]" true
file_must_contains "4. \[podnetworkchaos\]" true
echo "Cleaning up networkchaos"
kubectl delete -f delay.yaml
rm delay.yaml

echo "Run httpchaos"

cat <<EOF >delay.yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: HTTPChaos
metadata:
  name: web-show-http-delay
spec:
  mode: one
  selector:
    namespaces:
      - default
    labelSelectors:
      "app": "web-show"
  target: Request
  port: 8081
  path: "*"
  delay: "10ms"
  duration: "30s"
EOF
kubectl apply -f delay.yaml
sleep 3
echo "Checking chaosctl debug httpchaos"
./bin/chaosctl debug httpchaos web-show-http-delay >$log_file 2>&1
if [ $? -ne 0 ]; then
    echo "chaosctl debug httpchaos failed"
    code=1
fi
file_must_contains "\[Chaos\]: web-show-http-delay" true
file_must_contains "\[podhttpchaos\]" true
echo "Cleaning up httpchaos"
kubectl delete -f delay.yaml
rm delay.yaml

echo "Run iochaos"

cat <<EOF >delay.yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: IOChaos
metadata:
  name: web-show-io-delay
spec:
  action: latency
  mode: one
  selector:
    namespaces:
      - default
    labelSelectors:
      "app": "web-show"
  volumePath: /var/run/secrets/kubernetes.io/serviceaccount
  path: "/var/run/secrets/kubernetes.io/serviceaccount/**/*"
  delay: "10ms"
  percent: 50
  duration: "30s"
EOF
kubectl apply -f delay.yaml

# IOChaos only supports x86_64
# TODO: support IOChaos in aarch64, and remove this condition
if [[ "$(uname -m)" == "x86_64" ]]; then
  echo "Checking chaosctl debug iochaos"
  ./bin/chaosctl debug iochaos web-show-io-delay >$log_file 2>&1
  if [ $? -ne 0 ]; then
      echo "chaosctl debug iochaos failed"
      code=1
  fi
  file_must_contains "\[Chaos\]: web-show-io-delay" true
  file_must_contains "1. \[Mount Information\]" true
  file_must_contains "\[file descriptors of PID:" true
  file_must_contains "\[podiochaos\]" true
  echo "Cleaning up iochaos"
  kubectl delete -f delay.yaml
  rm delay.yaml
fi

echo "Run stresschaos"

cat <<EOF >stress.yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: StressChaos
metadata:
  name: web-show-memory-stress
spec:
  mode: one
  selector:
    namespaces:
      - default
    labelSelectors:
      "app": "web-show"
  stressors:
    memory:
      workers: 4
      size: "256MB"
  duration: "30s"
EOF
kubectl apply -f stress.yaml

echo "Checking chaosctl debug stresschaos"
./bin/chaosctl debug stresschaos web-show-memory-stress >$log_file 2>&1
if [ $? -ne 0 ]; then
    echo "chaosctl debug stresschaos failed"
    code=1
fi
file_must_contains "\[Chaos\]: web-show-memory-stress" true
file_must_contains "\[/proc/1/cgroup of web-show\]" true
echo "Cleaning up stresschaos"
kubectl delete -f stress.yaml
rm stress.yaml

rm $log_file
bash deploy.sh -d
rm deploy.sh
exit $code
