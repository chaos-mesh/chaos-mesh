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

echo $PATH

cur=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd $cur

echo "deploy deployments for test"
#kubectl apply -f https://raw.githubusercontent.com/chaos-mesh/apps/master/ping/busybox-statefulset.yaml

echo "create rbac and get token"

kubectl apply -f ./cluster-manager.yaml
kubectl apply -f ./cluster-viewer.yaml
kubectl apply -f ./busybox-manager.yaml
kubectl apply -f ./busybox-viewer.yaml

CLUSTER_MANAGER_TOKEN=`kubectl -n chaos-testing describe secret $(kubectl -n chaos-testing get secret | grep account-cluster-manager | awk '{print $1}') | grep "token:" | awk '{print $2}'`
CLUSTER_VIEWER_TOKEN=`kubectl -n chaos-testing describe secret $(kubectl -n chaos-testing get secret | grep account-busybox-viewer | awk '{print $1}') | grep "token:" | awk '{print $2}'`
BUSYBOX_MANAGER_TOKEN=`kubectl -n busybox describe secret $(kubectl -n busybox get secret | grep account-busybox-manager | awk '{print $1}') | grep "token:" | awk '{print $2}'`
BUSYBOX_VIEWER_TOKEN=`kubectl -n busybox describe secret $(kubectl -n busybox get secret | grep account-busybox-viewer | awk '{print $1}') | grep "token:" | awk '{print $2}'`

CLUSTER_VIEW_TOKEN_LIST=($CLUSTER_MANAGER_TOKEN $CLUSTER_VIEWER_TOKEN)
CLUSTER_VIEW_FORBIDDEN_TOKEN_LIST=($BUSYBOX_MANAGER_TOKEN $BUSYBOX_VIEWER_TOKEN)
BUSYBOX_MANAGE_TOKEN_LIST=($CLUSTER_MANAGER_TOKEN $BUSYBOX_MANAGER_TOKEN)
BUSYBOX_MANAGER_FORBIDDEN_TOKEN_LIST=($CLUSTER_VIEWER_TOKEN $BUSYBOX_VIEWER_TOKEN)
BUSYBOX_VIEW_TOKEN_LIST=($CLUSTER_MANAGER_TOKEN $CLUSTER_VIEWER_TOKEN $BUSYBOX_MANAGER_TOKEN $BUSYBOX_VIEWER_TOKEN)

CLUSTER_MANAGER_TOKEN_LIST=

EXP_JSON='{"name": "test", "namespace": "busybox", "scope": {"mode":"one", "namespace_selectors": ["busybox"]}, "target": {"kind": "NetworkChaos", "network_chaos": {"action": "delay", "delay": {"latency": "1ms"}}}}'

echo "***** create chaos experiments *****"

echo "viewer is forbidden to create experiments"
for(( i=0;i<${#BUSYBOX_MANAGER_FORBIDDEN_TOKEN_LIST[@]};i++)) do
    curl -X POST "localhost:2333/api/experiments/new" -H "Content-Type: application/json" -H "Authorization: Bearer ${BUSYBOX_MANAGER_FORBIDDEN_TOKEN_LIST[i]}" -d "${EXP_JSON}"  > create_exp.out
    check_contains "is forbidden" "create_exp.out"
done

#curl -X POST "localhost:2333/api/experiments/new" -H "Content-Type: application/json" -H "Authorization: Bearer ${CLUSTER_VIEWER_TOKEN}" -d "${EXP_JSON}"  > create_exp.out
#check_contains "is forbidden" "create_exp.out"
#curl -X POST "localhost:2333/api/experiments/new" -H "Content-Type: application/json" -H "Authorization: Bearer ${BUSYBOX_VIEWER_TOKEN}" -d "${EXP_JSON}"  > create_exp.out
#check_contains "is forbidden" "create_exp.out"
exit 0
echo "only manager can create experiments success"
curl -X POST "localhost:2333/api/experiments/new" -H "Content-Type: application/json" -H "Authorization: Bearer ${BUSYBOX_MANAGER_TOKEN}" -d "${EXP_JSON}" > create_exp.out
check_contains '"name":"test"' "create_exp.out"

echo "***** list chaos experiments *****"

echo "all token can list experiments under namespace busybox success"
curl -X GET "localhost:2333/api/experiments?namespace=busybox" -H "Authorization: Bearer ${BUSYBOX_MANAGER_TOKEN}" > list_exp.out
check_contains '"name":"test"' "list_exp.out"
EXP_UID=`cat list_exp.out | sed 's/.*\"uid\":\"\([0-9,a-z,-]*\)\".*/\1/g'`

echo "cluster manager and viewer can list all chaos experiments in the cluster"
curl -X GET "localhost:2333/api/experiments?namespace=busybox" -H "Authorization: Bearer ${CLUSTER_VIEWER_TOKEN}" > list_exp.out
check_contains '"name":"test"' "list_exp.out"

curl -X GET "localhost:2333/api/experiments?namespace=busybox" -H "Authorization: Bearer ${BUSYBOX_MANAGER_TOKEN}" > list_exp.out
check_contains '"name":"test"' "list_exp.out"


echo "get details of chaos experiments"
curl -X GET "localhost:2333/api/experiments/detail/${EXP_UID}" -H "Authorization: Bearer ${BUSYBOX_MANAGER_TOKEN}" > detail_exp.out
check_contains "Running" "detail_exp.out"
# {"kind":"","namespace":"busybox","name":"test","uid":"fc6bbe96-f251-47f2-a30c-bed1fe04bcf9","created":"2020-12-29T08:41:06Z","status":"Running","yaml":{"apiVersion":"","kind":"","metadata":{"name":"test","namespace":"busybox","labels":null,"annotations":{"experiment.chaos-mesh.org/pause":"false"}},"spec":{"action":"delay","mode":"one","value":"","selector":{"namespaces":["busybox"]},"delay":{"latency":"1ms","correlation":"0","jitter":"0ms"},"direction":"to"}}}

echo "get state"
curl -X GET "localhost:2333/api/experiments/state?namespace=busybox" -H "Authorization: Bearer ${BUSYBOX_MANAGER_TOKEN}" > state.out
check_contains '"Running":1' "state.out"
# {"Running":1,"Waiting":0,"Paused":0,"Failed":0,"Finished":0}


echo "pause chaos experiments"
curl -X PUT "localhost:2333/api/experiments/pause/${EXP_UID}?namespace=busybox" -H "Authorization: Bearer ${BUSYBOX_viewer_TOKEN}" > pause_exp.out
check_contains "is forbidden" "pause_exp.out"
# "is forbidden"

curl -X PUT "localhost:2333/api/experiments/pause/${EXP_UID}?namespace=busybox" -H "Authorization: Bearer ${BUSYBOX_MANAGER_TOKEN}" > pause_exp.out
check_contains "success" "pause_exp.out"
# {"status":"success"}

echo "restart chaos experiments"
curl -X PUT "localhost:2333/api/experiments/start/${EXP_UID}?namespace=busybox" -H "Authorization: Bearer ${BUSYBOX_viewer_TOKEN}" > restart_exp.out
check_contains "is forbidden" "restart_exp.out"
# "is forbidden"

curl -X PUT "localhost:2333/api/experiments/start/${EXP_UID}?namespace=busybox" -H "Authorization: Bearer ${BUSYBOX_MANAGER_TOKEN}" > restart_exp.out
check_contains "success" "restart_exp.out"
# {"status":"success"}                        

echo "update chaos experiments"
#  TODO


echo "delete chaos experiments"
curl -X DELETE "localhost:2333/api/experiments/${EXP_UID}" -H "Authorization: Bearer ${BUSYBOX_MANAGER_TOKEN}" > delete_exp.out
# check_contains 


echo "get pods"

echo "list archive chaos experiments"

curl -X GET "localhost:2333/api/archives?namespace=busybox" -H "Authorization: Bearer ${BUSYBOX_MANAGER_TOKEN}" > list_archives.out

# [{"uid":"1fce3d8b-44d8-4906-ae31-eeaefd443a60","kind":"NetworkChaos","namespace":"busybox","name":"test","action":"delay","start_time":"2020-12-29T08:23:22Z","finish_time":"2020-12-29T08:40:08.0139053Z"},{"uid":"fc6bbe96-f251-47f2-a30c-bed1fe04bcf9","kind":"NetworkChaos","namespace":"busybox","name":"test","action":"delay","start_time":"2020-12-29T08:41:06Z","finish_time":"2020-12-29T09:14:57.6823095Z"}]

echo "get detail of archive chaos experiment"
curl -X GET "localhost:2333/api/archives/detail?uid=${EXP_UID}&namespace=busybox" -H "Authorization: Bearer ${BUSYBOX_MANAGER_TOKEN}" > detail_archive.out
# {"uid":"1fce3d8b-44d8-4906-ae31-eeaefd443a60","kind":"NetworkChaos","namespace":"busybox","name":"test","action":"delay","start_time":"2020-12-29T08:23:22Z","finish_time":"2020-12-29T08:40:08.0139053Z","yaml":{"apiVersion":"chaos-mesh.org/v1alpha1","kind":"NetworkChaos","metadata":{"name":"test","namespace":"busybox","labels":null,"annotations":null},"spec":{"action":"delay","mode":"one","value":"","selector":{"namespaces":["busybox"]},"delay":{"latency":"1ms","correlation":"0","jitter":"0ms"},"direction":"to"}}}

echo "get report of archive chaos experiment"
curl -X GET "localhost:2333/api/archives/report?uid=${EXP_UID}&namespace=busybox" -H "Authorization: Bearer ${BUSYBOX_MANAGER_TOKEN}" > report_archive.out
# {"meta":{"uid":"1fce3d8b-44d8-4906-ae31-eeaefd443a60","kind":"NetworkChaos","namespace":"busybox","name":"test","action":"delay","start_time":"2020-12-29T08:23:22Z","finish_time":"2020-12-29T08:40:08.0139053Z"},"events":[{"id":2,"created_at":"2020-12-29T08:23:23.8662752Z","updated_at":"2020-12-29T08:40:07.9378684Z","deleted_at":null,"experiment":"test","namespace":"busybox","kind":"NetworkChaos","message":"","start_time":"2020-12-29T08:23:23Z","finish_time":"2020-12-29T08:40:07.9319133Z","duration":"","pods":[{"id":2,"created_at":"2020-12-29T08:23:23.8720842Z","updated_at":"2020-12-29T08:23:23.8720842Z","deleted_at":null,"event_id":2,"pod_ip":"10.244.0.38","pod_name":"busybox-0","namespace":"busybox","message":"This is a source pod.","action":"delay"}],"experiment_id":"1fce3d8b-44d8-4906-ae31-eeaefd443a60"}],"total_time":"16m46.0139053s","total_fault_time":"16m44.9319133s"}

echo "delete archive chaos experiment"
curl -X DELETE "localhost:2333/api/archives/${EXP_UID}?namespace=busybox" -H "Authorization: Bearer ${BUSYBOX_MANAGER_TOKEN}" > delete_archive.out

echo "list events"
curl -X GET "localhost:2333/api/events?namespace=busybox" -H "Authorization: Bearer ${BUSYBOX_MANAGER_TOKEN}" > list_event.out
# [{"id":2,"created_at":"2020-12-29T08:23:23.8662752Z","updated_at":"2020-12-29T08:40:07.9378684Z","deleted_at":null,"experiment":"test","namespace":"busybox","kind":"NetworkChaos","message":"","start_time":"2020-12-29T08:23:23Z","finish_time":"2020-12-29T08:40:07.9319133Z","duration":"","pods":[{"id":2,"created_at":"2020-12-29T08:23:23.8720842Z","updated_at":"2020-12-29T08:23:23.8720842Z","deleted_at":null,"event_id":2,"pod_ip":"10.244.0.38","pod_name":"busybox-0","namespace":"busybox","message":"This is a source pod.","action":"delay"}],"experiment_id":"1fce3d8b-44d8-4906-ae31-eeaefd443a60"},{"id":3,"created_at":"2020-12-29T08:41:06.4924757Z","updated_at":"2020-12-29T08:51:25.5754097Z","deleted_at":null,"experiment":"test","namespace":"busybox","kind":"NetworkChaos","message":"","start_time":"2020-12-29T08:41:06Z","finish_time":"2020-12-29T08:49:18Z","duration":"","pods":[{"id":3,"created_at":"2020-12-29T08:41:06.5018308Z","updated_at":"2020-12-29T08:41:06.5018308Z","deleted_at":null,"event_id":3,"pod_ip":"10.244.0.38","pod_name":"busybox-0","namespace":"busybox","message":"This is a source pod.","action":"delay"}],"experiment_id":"fc6bbe96-f251-47f2-a30c-bed1fe04bcf9"},{"id":4,"created_at":"2020-12-29T08:51:25.9029918Z","updated_at":"2020-12-29T09:02:34.4799263Z","deleted_at":null,"experiment":"test","namespace":"busybox","kind":"NetworkChaos","message":"","start_time":"2020-12-29T08:51:25Z","finish_time":"2020-12-29T09:02:34Z","duration":"","pods":[{"id":4,"created_at":"2020-12-29T08:51:25.9088941Z","updated_at":"2020-12-29T08:51:25.9088941Z","deleted_at":null,"event_id":4,"pod_ip":"10.244.0.38","pod_name":"busybox-0","namespace":"busybox","message":"This is a source pod.","action":"delay"}],"experiment_id":"fc6bbe96-f251-47f2-a30c-bed1fe04bcf9"}]
#EVENT_ID=

echo "get dry events"
curl -X GET "localhost:2333/api/events/dry?namespace=busybox" -H "Authorization: Bearer ${BUSYBOX_MANAGER_TOKEN}" > get_dry_event.out
# [{"id":2,"created_at":"2020-12-29T08:23:23.8662752Z","updated_at":"2020-12-29T08:40:07.9378684Z","deleted_at":null,"experiment":"test","namespace":"busybox","kind":"NetworkChaos","message":"","start_time":"2020-12-29T08:23:23Z","finish_time":"2020-12-29T08:40:07.9319133Z","duration":"","pods":null,"experiment_id":"1fce3d8b-44d8-4906-ae31-eeaefd443a60"},{"id":3,"created_at":"2020-12-29T08:41:06.4924757Z","updated_at":"2020-12-29T08:51:25.5754097Z","deleted_at":null,"experiment":"test","namespace":"busybox","kind":"NetworkChaos","message":"","start_time":"2020-12-29T08:41:06Z","finish_time":"2020-12-29T08:49:18Z","duration":"","pods":null,"experiment_id":"fc6bbe96-f251-47f2-a30c-bed1fe04bcf9"},{"id":4,"created_at":"2020-12-29T08:51:25.9029918Z","updated_at":"2020-12-29T09:02:34.4799263Z","deleted_at":null,"experiment":"test","namespace":"busybox","kind":"NetworkChaos","message":"","start_time":"2020-12-29T08:51:25Z","finish_time":"2020-12-29T09:02:34Z","duration":"","pods":null,"experiment_id":"fc6bbe96-f251-47f2-a30c-bed1fe04bcf9"}]

echo "get event by id"
curl -X GET "localhost:2333/api/events/get?id=1&namespace=busybox" -H "Authorization: Bearer ${BUSYBOX_MANAGER_TOKEN}" > get_event.out



#curl -X GET "localhost:2333/api/events?namespace=busybox" -H "Authorization: Bearer ${BUSYBOX_viewer_TOKEN}"
#curl -X GET "localhost:2333/api/events/get?namespace=busybox&id=1" -H "Authorization: Bearer ${BUSYBOX_MANAGER_TOKEN}"


