#!/usr/bin/env bash
# Copyright 2026 Chaos Mesh Authors.
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

# Manual node-level partial packet drop, driven through the chaos-daemon pod on
# the target node. Used to validate the gray-failure design before building it.
#
# USAGE
#   node-drop.sh apply  (--node NODE | --pod NS/NAME) --percent 40 --direction in|out|both
#                       [--include-host] [--no-exclude-local] [--ttl 600]
#                       [--exclude-namespaces chaos-mesh] [--host-exempt-ports 10250,31767]
#   node-drop.sh status (--node NODE | --pod NS/NAME)
#   node-drop.sh revert (--node NODE | --pod NS/NAME)
#
#   --node / --pod        Target node, or the node a pod runs on (e.g. a CockroachDB or NATS pod).
#   --percent             Drop probability per packet, in (0, 100].
#   --direction           in = INPUT, out = OUTPUT, both = both chains.
#   --include-host        Also drop in the host network namespace (kubelet, hostNetwork
#                         pods). TCP ports in --host-exempt-ports are never dropped there.
#   --no-exclude-local    Also drop traffic between pods/IPs on the same node.
#   --ttl                 Seconds until rules revert on their own (default 600).
#   --exclude-namespaces  Comma-separated namespaces whose pods are left alone
#                         (default: $DAEMON_NAMESPACE).
#
# ENVIRONMENT
#   DAEMON_NAMESPACE  namespace of chaos-daemon (default: chaos-mesh)
#   KUBECTL           kubectl command; use it to pin the cluster, e.g.
#                     KUBECTL="kubectl --context my-cluster"
#
# PREREQUISITES
#   - bash, and kubectl with exec access to pods in DAEMON_NAMESPACE.
#   - chaos-daemon running on the target node (label
#     app.kubernetes.io/component=chaos-daemon, container chaos-daemon). It must be
#     privileged with hostPID, which is the Helm chart default.
#   - The script is copied to /tmp/gf-node-drop.sh in the daemon pod on every run.
#
# HOW IT WORKS
#   In every non-hostNetwork pod netns on the node (and the host netns with
#   --include-host) it creates chains GF-IN / GF-OUT hooked first in INPUT / OUTPUT:
#   RETURN for loopback, RETURN for same-node IPs (ipset GF-LOCAL), and in the host
#   netns RETURN for exempt ports. Then: -m statistic --mode random --probability P -j DROP.
#   An auto-revert timer runs inside the daemon pod. If the daemon pod restarts, the
#   timer is lost; run "revert" by hand.
#
# RUNBOOK (CockroachDB incident)
#   export KUBECTL="kubectl --context <staging-context>"
#   POD=<crdb-ns>/<crdb-pod>
#
#   1. Baseline: note CockroachDB SQL p99, liveness heartbeat failures, leaseholder
#      moves and under-replicated ranges for ~10 minutes.
#   2. Sanity-check the loss rate. Pick a test pod on another node and one on the
#      target node (any image with ping), then:
#        node-drop.sh apply --pod $POD --percent 40 --direction in --ttl 300
#        kubectl exec <remote-pod> -- ping -q -c 500 -i 0.01 -W 1 <pod-ip-on-target-node>
#                                    expect ~40% loss (~64% with --direction both)
#        kubectl exec <local-pod>  -- ping -q -c 500 -i 0.01 -W 1 <pod-ip-on-target-node>
#                                    expect 0% loss (same-node traffic is excluded)
#        node-drop.sh revert --pod $POD
#   3. Incident runs, with at least 10 minutes of recovery between them:
#        node-drop.sh apply --pod $POD --percent 40 --direction in   --include-host --ttl 600
#        node-drop.sh apply --pod $POD --percent 40 --direction out  --include-host --ttl 600
#        node-drop.sh apply --pod $POD --percent 40 --direction both --include-host --ttl 600
#      During each run:
#        node-drop.sh status --pod $POD        DROP counters should keep increasing
#        kubectl get node <node> -w            watch for NotReady flaps
#   4. Clean up and confirm:
#        node-drop.sh revert --pod $POD
#        node-drop.sh status --pod $POD        expect "no gray-failure rules found"
#
#   Pass: the loss rate matches --percent, same-node traffic is unaffected, CockroachDB
#   degrades cluster-wide like the incident, and it recovers once the rules are reverted.
#   If pod-to-pod loss is ~0% on a Cilium cluster, the datapath bypasses pod netfilter
#   and the node-scope design needs revisiting.

set -euo pipefail

HERE=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REMOTE_SCRIPT="$HERE/node-drop-remote.sh"
REMOTE_PATH=/tmp/gf-node-drop.sh
DAEMON_NAMESPACE=${DAEMON_NAMESPACE:-chaos-mesh}
KUBECTL=${KUBECTL:-kubectl}

die() { echo "error: $*" >&2; exit 1; }
usage() { awk '/^$/ && !s { s = 1; next } s && /^#/ { sub(/^# ?/, ""); print; next } s { exit }' "${BASH_SOURCE[0]}"; exit 2; }

[[ $# -ge 1 ]] || usage
cmd=$1
shift
[[ $cmd == -h || $cmd == --help ]] && usage

node="" pod="" percent="" direction="" include_host=0 exclude_local=1 ttl=600
exclude_namespaces="$DAEMON_NAMESPACE" host_exempt_ports="10250,31767"
while [[ $# -gt 0 ]]; do
	case $1 in
	--node) node=$2; shift 2 ;;
	--pod) pod=$2; shift 2 ;;
	--percent) percent=$2; shift 2 ;;
	--direction) direction=$2; shift 2 ;;
	--include-host) include_host=1; shift ;;
	--no-exclude-local) exclude_local=0; shift ;;
	--ttl) ttl=$2; shift 2 ;;
	--exclude-namespaces) exclude_namespaces=$2; shift 2 ;;
	--host-exempt-ports) host_exempt_ports=$2; shift 2 ;;
	-h | --help) usage ;;
	*) die "unknown flag $1" ;;
	esac
done

if [[ -z $node ]]; then
	[[ -n $pod ]] || die "--node or --pod is required"
	node=$($KUBECTL get pod -n "${pod%%/*}" "${pod#*/}" -o jsonpath='{.spec.nodeName}')
	[[ -n $node ]] || die "pod $pod is not scheduled"
	echo "target node: $node (from pod $pod)"
fi

daemon=$($KUBECTL get pods -n "$DAEMON_NAMESPACE" \
	-l app.kubernetes.io/component=chaos-daemon \
	--field-selector "spec.nodeName=$node" \
	-o jsonpath='{.items[*].metadata.name}')
daemon=${daemon%% *}
[[ -n $daemon ]] || die "no chaos-daemon pod on node $node in namespace $DAEMON_NAMESPACE" \
	"(is the node tainted without a matching chaosDaemon.tolerations, or is DAEMON_NAMESPACE wrong?)"

remote() {
	$KUBECTL exec -i -n "$DAEMON_NAMESPACE" "$daemon" -c chaos-daemon -- "$@"
}

remote sh -c "cat > $REMOTE_PATH && chmod +x $REMOTE_PATH" <"$REMOTE_SCRIPT"

case $cmd in
apply)
	[[ -n $percent && -n $direction ]] || die "--percent and --direction are required"

	uids=() local_ips=()
	echo "pods on $node:"
	while IFS=$'\t' read -r ns name uid ip host_network; do
		[[ -n $ip ]] && [[ $host_network != true ]] && local_ips+=("$ip")
		if [[ $host_network == true ]]; then
			printf '  %-6s %s/%s (hostNetwork)\n' skip "$ns" "$name"
		elif [[ ",$exclude_namespaces," == *",$ns,"* ]]; then
			printf '  %-6s %s/%s (excluded namespace)\n' skip "$ns" "$name"
		else
			printf '  %-6s %s/%s %s\n' target "$ns" "$name" "$ip"
			uids+=("$uid")
		fi
	done < <($KUBECTL get pods -A \
		--field-selector "spec.nodeName=$node,status.phase=Running" \
		-o jsonpath='{range .items[*]}{.metadata.namespace}{"\t"}{.metadata.name}{"\t"}{.metadata.uid}{"\t"}{.status.podIP}{"\t"}{.spec.hostNetwork}{"\n"}{end}')

	while read -r ip; do
		[[ -n $ip ]] && local_ips+=("$ip")
	done < <($KUBECTL get node "$node" \
		-o jsonpath='{range .status.addresses[?(@.type=="InternalIP")]}{.address}{"\n"}{end}')

	[[ ${#uids[@]} -gt 0 || $include_host == 1 ]] || die "nothing to target on $node"

	args=(apply --percent "$percent" --direction "$direction" --ttl "$ttl"
		--uids "$(IFS=,; echo "${uids[*]}")"
		--host-exempt-ports "$host_exempt_ports")
	if [[ $exclude_local == 1 ]]; then
		args+=(--local-ips "$(IFS=,; echo "${local_ips[*]}")")
	fi
	if [[ $include_host == 1 ]]; then
		args+=(--include-host)
	fi
	remote bash "$REMOTE_PATH" "${args[@]}"
	;;
status | revert)
	remote bash "$REMOTE_PATH" "$cmd"
	;;
*)
	usage
	;;
esac
