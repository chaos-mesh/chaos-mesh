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

set -eu

cur=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd $cur

echo "deploy deployments for test"
kubectl apply -f https://raw.githubusercontent.com/chaos-mesh/apps/master/ping/busybox-statefulset.yaml

# wait pod status to running
for ((k=0; k<30; k++)); do
    kubectl get pods --namespace busybox > pods.status
    cat pods.status

    run_num=`grep Running pods.status | wc -l`
    pod_num=$((`cat pods.status | wc -l` - 1))
    if [ $run_num == $pod_num ]; then
        break
    fi

    sleep 1
done

echo "create rbac and get token"

kubectl apply -f ./cluster-manager.yaml
kubectl apply -f ./cluster-viewer.yaml
kubectl apply -f ./busybox-manager.yaml
kubectl apply -f ./busybox-viewer.yaml

CLUSTER_MANAGER_TOKEN=`kubectl -n chaos-testing describe secret $(kubectl -n chaos-testing get secret | grep account-cluster-manager | awk '{print $1}') | grep "token:" | awk '{print $2}'`
CLUSTER_VIEWER_TOKEN=`kubectl -n chaos-testing describe secret $(kubectl -n chaos-testing get secret | grep account-cluster-viewer | awk '{print $1}') | grep "token:" | awk '{print $2}'`
BUSYBOX_MANAGER_TOKEN=`kubectl -n busybox describe secret $(kubectl -n busybox get secret | grep account-busybox-manager | awk '{print $1}') | grep "token:" | awk '{print $2}'`
BUSYBOX_VIEWER_TOKEN=`kubectl -n busybox describe secret $(kubectl -n busybox get secret | grep account-busybox-viewer | awk '{print $1}') | grep "token:" | awk '{print $2}'`

BUSYBOX_MANAGER_TOKEN_LIST=($BUSYBOX_MANAGER_TOKEN)
CLUSTER_MANAGER_TOKEN_LIST=($CLUSTER_MANAGER_TOKEN)

CLUSTER_VIEW_TOKEN_LIST=($CLUSTER_MANAGER_TOKEN $CLUSTER_VIEWER_TOKEN)
CLUSTER_VIEW_FORBIDDEN_TOKEN_LIST=($BUSYBOX_MANAGER_TOKEN $BUSYBOX_VIEWER_TOKEN)
CLUSTER_MANAGER_FORBIDDEN_TOKEN_LIST=($CLUSTER_VIEWER_TOKEN $BUSYBOX_MANAGER_TOKEN $BUSYBOX_VIEWER_TOKEN)
BUSYBOX_MANAGE_TOKEN_LIST=($CLUSTER_MANAGER_TOKEN $BUSYBOX_MANAGER_TOKEN)
BUSYBOX_MANAGER_FORBIDDEN_TOKEN_LIST=($CLUSTER_VIEWER_TOKEN $BUSYBOX_VIEWER_TOKEN)
BUSYBOX_VIEW_TOKEN_LIST=($CLUSTER_MANAGER_TOKEN $CLUSTER_VIEWER_TOKEN $BUSYBOX_MANAGER_TOKEN $BUSYBOX_VIEWER_TOKEN)

EXP_JSON='{"name": "ci-test", "namespace": "busybox", "scope": {"mode":"one", "namespaces": ["busybox"]}, "target": {"kind": "NetworkChaos", "network_chaos": {"action": "delay", "delay": {"latency": "1ms"}}}}'
UPDATE_EXP_JSON='{"apiVersion": "chaos-mesh.org/v1alpha1", "kind": "NetworkChaos", "metadata": {"name": "ci-test", "namespace": "busybox"}, "spec": {"action": "delay", "latency": "2ms", "mode": "one"}}'

function REQUEST() {
    declare -a TOKEN_LIST=("${!1}")
    METHOD=$2
    URL=$3
    LOG=$4
    MESSAGE=$5

    for(( i=0;i<${#TOKEN_LIST[@]};i++)) do
        echo "$i. use token ${TOKEN_LIST[i]: 0: 10}...${TOKEN_LIST[i]: 0-10} to send $METHOD request to $URL, and save log in $LOG, log should contains '$MESSAGE'"
        if [ "$METHOD" == "POST" ]; then
            curl -X $METHOD "localhost:2333$URL" -H "Content-Type: application/json" -H "Authorization: Bearer ${TOKEN_LIST[i]}" -d "${EXP_JSON}" > $LOG
        elif [ "$METHOD" == "PUT" ]; then
            curl -X $METHOD "localhost:2333$URL" -H "Content-Type: application/json" -H "Authorization: Bearer ${TOKEN_LIST[i]}" -d "${UPDATE_EXP_JSON}" > $LOG
        else
            curl -X $METHOD "localhost:2333$URL" -H "Authorization: Bearer ${TOKEN_LIST[i]}" > $LOG
        fi
        check_contains "$MESSAGE" $LOG
    done
}

echo "***** create chaos experiments *****"

echo "viewer is forbidden to create experiments"
REQUEST BUSYBOX_MANAGER_FORBIDDEN_TOKEN_LIST[@] "POST" "/api/experiments/new" "create_exp.out" "is forbidden"

echo "only manager can create experiments success"
# here just use busybox manager because experiment can be created only one time
REQUEST BUSYBOX_MANAGER_TOKEN_LIST[@] "POST" "/api/experiments/new" "create_exp.out" '"name":"ci-test"'


echo "***** list chaos experiments *****"

echo "all token can list experiments under namespace busybox"
REQUEST BUSYBOX_VIEW_TOKEN_LIST[@] "GET" "/api/experiments?namespace=busybox" "list_exp.out" '"name":"ci-test"'

EXP_UID=`cat list_exp.out | sed 's/.*\"uid\":\"\([0-9,a-z,-]*\)\".*/\1/g'`

echo "cluster manager and viewer can list all chaos experiments in the cluster"
REQUEST CLUSTER_VIEW_TOKEN_LIST[@] "GET" "/api/experiments" "list_exp.out" '"name":"ci-test"'

echo "busybox manager and viewer is forbidden to list chaos experiments in the cluster or other namespace"
REQUEST CLUSTER_VIEW_FORBIDDEN_TOKEN_LIST[@] "GET" "/api/experiments" "list_exp.out" "is forbidden"
REQUEST CLUSTER_VIEW_FORBIDDEN_TOKEN_LIST[@] "GET" "/api/experiments?namespace=default" "list_exp.out" "is forbidden"


#echo "***** get details of chaos experiments *****"
#
#echo "all token can view the experiments under namespace busybox"
#REQUEST BUSYBOX_VIEW_TOKEN_LIST[@] "GET" "/api/experiments/detail/${EXP_UID}?namespace=busybox" "exp_detail.out" "Running"
#
#
#echo "***** get state of chaos experiments *****"
#
#echo "all token can get the state of experiments under namespace busybox"
#REQUEST BUSYBOX_VIEW_TOKEN_LIST[@] "GET" "/api/experiments/state?namespace=busybox" "exp_state.out" "Running"
#
#echo "cluster manager and viewer can get the state of experiments in the cluster"
#REQUEST CLUSTER_VIEW_TOKEN_LIST[@] "GET" "/api/experiments/state" "exp_state.out" "Running"
#
#echo "busybox manager and viewer is forbidden to get the state of experiments in the cluster or other namespace"
#REQUEST CLUSTER_VIEW_FORBIDDEN_TOKEN_LIST[@] "GET" "/api/experiments/state" "exp_state.out" "is forbidden"
#REQUEST CLUSTER_VIEW_FORBIDDEN_TOKEN_LIST[@] "GET" "/api/experiments/state?namespace=default" "exp_state.out" "is forbidden"


echo "***** pause chaos experiments *****"

echo "viewer is forbidden to pause experiments"
REQUEST BUSYBOX_MANAGER_FORBIDDEN_TOKEN_LIST[@] "PUT" "/api/experiments/pause/${EXP_UID}?namespace=busybox" "pause_exp.out" "is forbidden"

echo "only manager can pause experiments"
REQUEST BUSYBOX_MANAGE_TOKEN_LIST[@] "PUT" "/api/experiments/pause/${EXP_UID}?namespace=busybox" "pause_exp.out" "success"

echo "***** restart chaos experiments *****"

echo "viewer is forbidden to restart experiments"
REQUEST BUSYBOX_MANAGER_FORBIDDEN_TOKEN_LIST[@] "PUT" "/api/experiments/start/${EXP_UID}?namespace=busybox" "restart_exp.out" "is forbidden"

echo "only manager can pause experiments"
REQUEST BUSYBOX_MANAGE_TOKEN_LIST[@] "PUT" "/api/experiments/start/${EXP_UID}?namespace=busybox" "restart_exp.out" "success"


# As we are discussing whether we should provide the ability to modify a chaos while running, these tests are removed now
# echo "***** update chaos experiments *****"

# echo "viewer is forbidden to update experiments"
# REQUEST BUSYBOX_MANAGER_FORBIDDEN_TOKEN_LIST[@] "PUT" "/api/experiments/update" "update_exp.out" "is forbidden"

# echo "only manager can update experiments"
# REQUEST BUSYBOX_MANAGE_TOKEN_LIST[@] "PUT" "/api/experiments/update" "update_exp.out" '"name":"ci-test"'


echo "***** delete chaos experiments *****"

echo "viewer is forbidden to delete experiments"
REQUEST BUSYBOX_MANAGER_FORBIDDEN_TOKEN_LIST[@] "DELETE" "/api/experiments/${EXP_UID}" "delete_exp.out" "is forbidden"

echo "only manager can delete experiments success"
# here just use cluster manager because experiment can be delete only one time
REQUEST CLUSTER_MANAGER_TOKEN_LIST[@] "DELETE" "/api/experiments/${EXP_UID}" "delete_exp.out" "success"


echo "***** list events *****"

echo "all token can list events under namespace busybox"
REQUEST BUSYBOX_VIEW_TOKEN_LIST[@] "GET" "/api/events?namespace=busybox" "list_event.out" "ci-test"

EVENT_ID=`cat list_event.out | sed 's/.*\"id\":\([0-9]*\),.*/\1/g'`

echo "cluster manager and viewer can list events in the cluster"
REQUEST CLUSTER_VIEW_TOKEN_LIST[@] "GET" "/api/events" "list_event.out" "ci-test"

echo "busybox manager and viewer is forbidden to list events in the cluster or other namespace"
REQUEST CLUSTER_VIEW_FORBIDDEN_TOKEN_LIST[@] "GET" "/api/events" "list_event.out" "can't list"
REQUEST CLUSTER_VIEW_FORBIDDEN_TOKEN_LIST[@] "GET" "/api/events?namespace=default" "list_event.out" "can't list"


#echo "***** list dry events *****"
#
#echo "all token can list dry events under namespace busybox"
#REQUEST BUSYBOX_VIEW_TOKEN_LIST[@] "GET" "/api/events/dry?namespace=busybox" "list_dry_event.out" "ci-test"
#
#echo "cluster manager and viewer can list dry events in the cluster"
#REQUEST CLUSTER_VIEW_TOKEN_LIST[@] "GET" "/api/events/dry" "list_dry_event.out" "ci-test"
#
#echo "busybox manager and viewer is forbidden to list dry events in the cluster or other namespace"
#REQUEST CLUSTER_VIEW_FORBIDDEN_TOKEN_LIST[@] "GET" "/api/events/dry" "list_dry_event.out" "can't list"
#REQUEST CLUSTER_VIEW_FORBIDDEN_TOKEN_LIST[@] "GET" "/api/events/dry?namespace=default" "list_dry_event.out" "can't list"


echo "***** get event by id *****"

echo "all token can get event under namespace busybox"
REQUEST BUSYBOX_VIEW_TOKEN_LIST[@] "GET" "/api/events/get?id=$EVENT_ID&namespace=busybox" "get_event.out" "ci-test"

echo "cluster manager and viewer can get event in the cluster"
REQUEST CLUSTER_VIEW_TOKEN_LIST[@] "GET" "/api/events/get?id=$EVENT_ID" "get_event.out" "ci-test"

echo "busybox manager and viewer is forbidden to get event in the cluster or other namespace"
REQUEST CLUSTER_VIEW_FORBIDDEN_TOKEN_LIST[@] "GET" "/api/events/get?id=$EVENT_ID" "get_event.out" "can't list"
REQUEST CLUSTER_VIEW_FORBIDDEN_TOKEN_LIST[@] "GET" "/api/events/get?id=$EVENT_ID&namespace=default" "get_event.out" "can't list"


echo "***** list archive chaos experiments *****"

echo "all token can list archive experiments under namespace busybox"
REQUEST BUSYBOX_VIEW_TOKEN_LIST[@] "GET" "/api/archives?namespace=busybox" "list_archives.out" '"name":"ci-test"'

echo "cluster manager and viewer can list archive experiments in the cluster"
REQUEST CLUSTER_VIEW_TOKEN_LIST[@] "GET" "/api/archives" "list_archives.out" '"name":"ci-test"'

echo "busybox manager and viewer is forbidden to list archive experiments in the cluster or other namespace"
REQUEST CLUSTER_VIEW_FORBIDDEN_TOKEN_LIST[@] "GET" "/api/archives" "list_archives.out" "can't list"
REQUEST CLUSTER_VIEW_FORBIDDEN_TOKEN_LIST[@] "GET" "/api/archives?namespace=default" "list_archives.out" "can't list"


echo "***** get detail of archive chaos experiment *****"

echo "all token can get the details of archive experiments under namespace busybox"
REQUEST BUSYBOX_VIEW_TOKEN_LIST[@] "GET" "/api/archives/detail?uid=${EXP_UID}&namespace=busybox" "detail_archives.out" '"name":"ci-test"'

echo "cluster manager and viewer can get the details of archive experiments in the cluster"
REQUEST CLUSTER_VIEW_TOKEN_LIST[@] "GET" "/api/archives/detail?uid=${EXP_UID}" "detail_archives.out" '"name":"ci-test"'

echo "busybox manager and viewer is forbidden to get the details of archive experiments in the cluster or other namespace"
REQUEST CLUSTER_VIEW_FORBIDDEN_TOKEN_LIST[@] "GET" "/api/archives/detail?uid=${EXP_UID}" "detail_archives.out" "can't list"
REQUEST CLUSTER_VIEW_FORBIDDEN_TOKEN_LIST[@] "GET" "/api/archives/detail?uid=${EXP_UID}&namespace=default" "detail_archives.out" "can't list"


#echo "***** get report of archive chaos experiment *****"
#
#echo "all token can get the report of archive experiments under namespace busybox"
#REQUEST BUSYBOX_VIEW_TOKEN_LIST[@] "GET" "/api/archives/report?uid=${EXP_UID}&namespace=busybox" "report_archives.out" '"name":"ci-test"'
#
#echo "cluster manager and viewer can get the report of archive experiments in the cluster"
#REQUEST CLUSTER_VIEW_TOKEN_LIST[@] "GET" "/api/archives/report?uid=${EXP_UID}" "report_archives.out" '"name":"ci-test"'
#
#echo "busybox manager and viewer is forbidden to get the report of archive experiments in the cluster or other namespace"
#REQUEST CLUSTER_VIEW_FORBIDDEN_TOKEN_LIST[@] "GET" "/api/archives/report?uid=${EXP_UID}" "report_archives.out" "can't list"
#REQUEST CLUSTER_VIEW_FORBIDDEN_TOKEN_LIST[@] "GET" "/api/archives/report?uid=${EXP_UID}&namespace=default" "report_archives.out" "can't list"
#
#
#echo "***** delete archive chaos experiment *****"
#
#echo "viewer is forbidden to delete archive experiments"
#REQUEST BUSYBOX_MANAGER_FORBIDDEN_TOKEN_LIST[@] "DELETE" "/api/archives/${EXP_UID}?namespace=busybox" "delete_archives.out" "can't"
#
#echo "only manager can delete archive experiments success"
## here use one manager token to delete it
#REQUEST BUSYBOX_MANAGER_TOKEN_LIST[@] "DELETE" "/api/archives/${EXP_UID}?namespace=busybox" "delete_archives.out" "success"


echo "***** test webhook authority ******"

EXP_JSON='{"name": "ci-test2", "namespace": "busybox", "scope": {"mode": "one", "namespaces": ["busybox"]}, "target": {"kind": "NetworkChaos", "network_chaos": {"direction": "both", "target_scope": {"namespaces": ["chaos-testing"], "mode": "one"}, "action": "delay", "delay": {"latency": "1ms"}}}}'
UPDATE_EXP_JSON='{"apiVersion": "chaos-mesh.org/v1alpha1", "kind": "NetworkChaos", "metadata": {"name": "ci-test2", "namespace": "busybox"}, "spec": {"direction": "both", "target": {"selector": {"namespaces": ["chaos-testing", "default" ]}, "mode": "one"}, "action": "delay", "latency": "2ms", "mode": "one"}}'

# create experiment require the privileges of namespace busybox and chaos-testing, so only cluster manager can create exp success
REQUEST CLUSTER_MANAGER_FORBIDDEN_TOKEN_LIST[@] "POST" "/api/experiments/new" "create_exp.out" 'is forbidden'

REQUEST CLUSTER_MANAGER_TOKEN_LIST[@] "POST" "/api/experiments/new" "create_exp.out" '"name":"ci-test2"'

# update the experiment require the privileges of namespace busybox, chaos-testing and default, so only cluster manager can update exp success

# As we are discussing whether we should provide the ability to modify a chaos while running, these tests are removed now
# REQUEST CLUSTER_MANAGER_FORBIDDEN_TOKEN_LIST[@] "PUT" "/api/experiments/update" "update_exp.out" "is forbidden"

# REQUEST CLUSTER_MANAGER_TOKEN_LIST[@] "PUT" "/api/experiments/update" "update_exp.out" '"name":"ci-test2"'

# delete experiment
kubectl delete networkchaos.chaos-mesh.org ci-test2 -n busybox

echo "pass the dashboard authority test!"
