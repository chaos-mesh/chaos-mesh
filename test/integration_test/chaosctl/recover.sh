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

function must_contains() {
  message=$1
  substring=$2
  match=""
  if [ "$3" = "false" ]; then
      match="-v"
  fi

  echo $message | grep $match "$substring"
  if [ "$?" != "0" ]; then
      echo "'$substring' not found in '$message'"
      exit 1
  fi
}

function get_high_latency() {
    # the log looks like `64 bytes from 10.244.0.9: seq=0 ttl=63 time=0.240 ms`
    # get the latency from log
    latencies=$(kubectl exec busybox-0 -i -n busybox -- ping -c 10 busybox-1.busybox.busybox.svc | grep "bytes from"  | sed 's/.*time=\([0-9]*\).[0-9]* ms.*/\1/g')
    high_latency_num=0
    for latency in $latencies
    do
        if [[ "$latency" -ge "10" ]]; then
            high_latency_num=`expr $high_latency_num + 1`
        fi
    done
    echo $high_latency_num
}

echo "Deploy web-show for testing"
curl https://mirrors.chaos-mesh.org/v1.1.2/web-show/deploy.sh | bash -s

echo "Deploy busyboxplus for test"
kubectl run busyboxplus --image=radial/busyboxplus:curl -- sleep 3600

echo "deploy busybox for test"
kubectl apply -f https://raw.githubusercontent.com/chaos-mesh/apps/master/ping/busybox-statefulset.yaml

echo "wait pods status to running"
for ((k=0; k<30; k++)); do
    webshow_num=$(kubectl get pods -l app=web-show | grep "Running" | wc -l)
    busyboxplus_num=$(kubectl get pods busyboxplus | grep "Running" | wc -l)
    busybox_num=$(kubectl get pods --namespace busybox | grep "Running" | wc -l)
    if [ $webshow_num == 1 ] && [ $busyboxplus_num == 1 ] && [ $busybox_num == 2 ]; then
        break
    fi
    sleep 1
done

echo "Confirm web-show works well"
must_contains "$(kubectl exec busyboxplus -- sh -c "curl -I web-show.default:8081")" "HTTP/1.1 200 OK" true

echo "Run networkchaos"

cat <<EOF >delay.yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: network-delay
  namespace: busybox
spec:
  action: delay
  mode: all
  selector:
    pods:
      busybox:
        - busybox-0
  delay:
    latency: "10ms"
  duration: "5s"
  direction: to
  target:
    selector:
      pods:
        busybox:
          - busybox-1
    mode: all
EOF
kubectl apply -f delay.yaml

echo "Confirm networkchaos works well"
sleep 1 # TODO: better way to wait for chaos being injected

high_latency_num=$(get_high_latency)

# about half of the latency should be greater than 10ms
if [[ "$high_latency_num" -lt "3" ]] || [[ "$high_latency_num" -gt "6" ]]; then
    echo "the chaos dosen't work as expect"
    exit 1
fi

echo "Recover networkchaos"
./bin/chaosctl recover networkchaos busybox-0 -n busybox

echo "Confirm httpchaos recovered"
high_latency_num=$(get_high_latency)

if [[ "$high_latency_num" -gt "0" ]]; then
    echo "the httpchaos dosen't recover"
    exit 1
fi

echo "Cleaning up networkchaos"
kubectl delete -f delay.yaml
rm delay.yaml

echo "Run httpchaos"

cat <<EOF >replace.yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: HTTPChaos
metadata:
  name: web-show-http-replace
spec:
  mode: one # the mode to run chaos action; supported modes are one/all/fixed/fixed-percent/random-max-percent
  selector: # pods where to inject chaos actions
    namespaces:
      - default
    labelSelectors:
      "app": "web-show"  # the label of the pod for chaos injection
  target: Response
  port: 8081
  path: "*"
  replace:
    code: 404
EOF
kubectl apply -f replace.yaml

echo "Confirm httpchaos works well"
sleep 1 # TODO: better way to wait for chaos being injected
must_contains "$(kubectl exec busyboxplus -- sh -c "curl -I web-show.default:8081")" "HTTP/1.1 404 Not Found" true

echo "Recover httpchaos"
./bin/chaosctl recover httpchaos -l app=web-show

echo "Confirm httpchaos recovered"
must_contains "$(kubectl exec busyboxplus -- sh -c "curl -I web-show.default:8081")" "HTTP/1.1 200 OK" true

echo "Cleaning up httpchaos"
kubectl delete -f replace.yaml
rm replace.yaml

kubectl delete pod busyboxplus
curl https://mirrors.chaos-mesh.org/v1.1.2/web-show/deploy.sh | bash -s -- -d
kubectl delete -f https://raw.githubusercontent.com/chaos-mesh/apps/master/ping/busybox-statefulset.yaml
