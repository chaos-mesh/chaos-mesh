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

# Runs inside a chaos-daemon pod (privileged, hostPID). Applies, reports and
# reverts probabilistic packet drop in the network namespaces of selected pods
# on this node, and optionally in the host network namespace.
#
#   node-drop-remote.sh apply --percent 40 --direction in|out|both \
#       --uids UID,UID [--local-ips IP,IP] [--include-host] \
#       [--host-exempt-ports 10250,31767] [--ttl 600]
#   node-drop-remote.sh status
#   node-drop-remote.sh revert

set -euo pipefail

CHAIN_IN=GF-IN
CHAIN_OUT=GF-OUT
LOCAL_SET=GF-LOCAL
SELF=/tmp/gf-node-drop.sh
TTL_PID_FILE=/tmp/gf-node-drop.ttl.pid
UID_RE='pod([0-9a-f]{8}[-_][0-9a-f]{4}[-_][0-9a-f]{4}[-_][0-9a-f]{4}[-_][0-9a-f]{12})'

log() { echo "[$(hostname)] $*" >&2; }

ipt() {
	local pid=$1
	shift
	nsenter --net="/proc/${pid}/ns/net" -- iptables -w "$@"
}

ips() {
	local pid=$1
	shift
	nsenter --net="/proc/${pid}/ns/net" -- ipset "$@"
}

netns_of() { readlink "/proc/$1/ns/net" 2>/dev/null; }

# Prints "uid pid" for one process per pod netns found on the host.
scan_pods() {
	local f content uid pid
	declare -A seen=()
	for f in /proc/[0-9]*/cgroup; do
		{ content=$(<"$f"); } 2>/dev/null || continue
		[[ $content =~ $UID_RE ]] || continue
		uid=${BASH_REMATCH[1]//_/-}
		[[ -n ${seen[$uid]:-} ]] && continue
		pid=${f#/proc/}
		pid=${pid%/cgroup}
		seen[$uid]=1
		echo "$uid $pid"
	done
}

has_chains() {
	ipt "$1" -S "$CHAIN_IN" >/dev/null 2>&1 || ipt "$1" -S "$CHAIN_OUT" >/dev/null 2>&1
}

# Prints "label pid" for one process per distinct netns that carries our chains.
# Every process is considered, not only the ones matched to a pod, so no netns
# with leftover chains can be missed.
scan_injected() {
	local host_ns f pid ns content label
	declare -A seen=()
	host_ns=$(netns_of 1)
	has_chains 1 && echo "host 1"
	seen[$host_ns]=1
	for f in /proc/[0-9]*/ns/net; do
		pid=${f#/proc/}
		pid=${pid%%/*}
		ns=$(netns_of "$pid") || continue
		[[ -n ${seen[$ns]:-} ]] && continue
		seen[$ns]=1
		has_chains "$pid" || continue
		label="pid-$pid"
		{ content=$(<"/proc/$pid/cgroup"); } 2>/dev/null || true
		[[ ${content:-} =~ $UID_RE ]] && label=${BASH_REMATCH[1]//_/-}
		echo "$label $pid"
	done
}

# build_chain PID CHAIN HOOK IFACE_FLAG SET_DIR PROBABILITY EXCLUDE_LOCAL EXEMPT_PORTS
build_chain() {
	local pid=$1 chain=$2 hook=$3 iface=$4 setdir=$5 prob=$6 exclude_local=$7 ports=$8

	ipt "$pid" -N "$chain" 2>/dev/null || true
	ipt "$pid" -F "$chain"
	ipt "$pid" -A "$chain" "$iface" lo -j RETURN
	if [[ $exclude_local == 1 ]]; then
		ipt "$pid" -A "$chain" -m set --match-set "$LOCAL_SET" "$setdir" -j RETURN
	fi
	if [[ -n $ports ]]; then
		ipt "$pid" -A "$chain" -p tcp -m multiport --dports "$ports" -j RETURN
		ipt "$pid" -A "$chain" -p tcp -m multiport --sports "$ports" -j RETURN
	fi
	ipt "$pid" -A "$chain" -m statistic --mode random --probability "$prob" -j DROP
	ipt "$pid" -C "$hook" -j "$chain" 2>/dev/null || ipt "$pid" -I "$hook" 1 -j "$chain"
}

inject_netns() {
	local pid=$1 direction=$2 prob=$3 local_ips=$4 ports=$5
	local exclude_local=0 ip

	if [[ -n $local_ips ]]; then
		exclude_local=1
		ips "$pid" create "$LOCAL_SET" hash:net -exist
		ips "$pid" flush "$LOCAL_SET"
		for ip in ${local_ips//,/ }; do
			ips "$pid" add "$LOCAL_SET" "$ip" -exist
		done
	fi

	if [[ $direction == in || $direction == both ]]; then
		build_chain "$pid" "$CHAIN_IN" INPUT -i src "$prob" "$exclude_local" "$ports"
	fi
	if [[ $direction == out || $direction == both ]]; then
		build_chain "$pid" "$CHAIN_OUT" OUTPUT -o dst "$prob" "$exclude_local" "$ports"
	fi
}

clean_netns() {
	local pid=$1 hook chain
	for pair in "INPUT:$CHAIN_IN" "OUTPUT:$CHAIN_OUT"; do
		hook=${pair%%:*}
		chain=${pair#*:}
		while ipt "$pid" -D "$hook" -j "$chain" 2>/dev/null; do :; done
		ipt "$pid" -F "$chain" 2>/dev/null || true
		ipt "$pid" -X "$chain" 2>/dev/null || true
	done
	ips "$pid" destroy "$LOCAL_SET" 2>/dev/null || true
}

cancel_ttl() {
	# The timer may exec straight into "revert", so it must not signal itself.
	if [[ -n ${GF_FROM_TTL:-} ]]; then
		rm -f "$TTL_PID_FILE"
	elif [[ -f $TTL_PID_FILE ]]; then
		kill -- "-$(<"$TTL_PID_FILE")" 2>/dev/null || true
		rm -f "$TTL_PID_FILE"
	fi
}

cmd_apply() {
	local percent="" direction="" uids="" local_ips="" include_host=0 ports="10250,31767" ttl=600
	while [[ $# -gt 0 ]]; do
		case $1 in
		--percent) percent=$2; shift 2 ;;
		--direction) direction=$2; shift 2 ;;
		--uids) uids=$2; shift 2 ;;
		--local-ips) local_ips=$2; shift 2 ;;
		--include-host) include_host=1; shift ;;
		--host-exempt-ports) ports=$2; shift 2 ;;
		--ttl) ttl=$2; shift 2 ;;
		*) log "unknown flag $1"; exit 2 ;;
		esac
	done
	[[ $direction =~ ^(in|out|both)$ ]] || { log "--direction must be in, out or both"; exit 2; }
	[[ $percent =~ ^[0-9]+(\.[0-9]+)?$ ]] || { log "--percent must be a number"; exit 2; }

	local prob
	prob=$(awk -v p="$percent" 'BEGIN { if (p <= 0 || p > 100) exit 1; printf "%.4f", p / 100 }') ||
		{ log "--percent must be in (0, 100]"; exit 2; }

	declare -A want=()
	local uid pid
	for uid in ${uids//,/ }; do want[$uid]=1; done

	local host_ns injected=0
	host_ns=$(netns_of 1)
	while read -r uid pid; do
		[[ -n ${want[$uid]:-} ]] || continue
		if [[ $(netns_of "$pid") == "$host_ns" ]]; then
			log "skip $uid: shares host netns"
			continue
		fi
		inject_netns "$pid" "$direction" "$prob" "$local_ips" ""
		log "injected pod $uid (pid $pid)"
		unset "want[$uid]"
		injected=$((injected + 1))
	done < <(scan_pods)

	for uid in "${!want[@]}"; do log "WARN no process found for pod $uid"; done

	if [[ $include_host == 1 ]]; then
		inject_netns 1 "$direction" "$prob" "$local_ips" "$ports"
		log "injected host netns (exempt tcp ports: $ports)"
	fi

	cancel_ttl
	setsid bash -c "sleep $ttl; GF_FROM_TTL=1 $SELF revert" >/tmp/gf-node-drop.ttl.log 2>&1 </dev/null &
	echo $! >"$TTL_PID_FILE"
	log "applied: ${percent}% ${direction}, ${injected} pod netns, host=${include_host}, auto-revert in ${ttl}s"
}

cmd_status() {
	local label pid chain
	local found=0
	while read -r label pid; do
		found=1
		echo "=== $label (pid $pid)"
		for chain in "$CHAIN_IN" "$CHAIN_OUT"; do
			ipt "$pid" -L "$chain" -v -n -x 2>/dev/null || true
		done
	done < <(scan_injected)
	[[ $found == 1 ]] || echo "no gray-failure rules found on this node"
	if [[ -f $TTL_PID_FILE ]] && kill -0 "$(<"$TTL_PID_FILE")" 2>/dev/null; then
		echo "auto-revert timer running (pid $(<"$TTL_PID_FILE"))"
	elif [[ $found == 1 ]]; then
		echo "WARNING: rules are present but no auto-revert timer is running; run revert"
	fi
}

cmd_revert() {
	local label pid n=0
	while read -r label pid; do
		clean_netns "$pid"
		log "reverted $label (pid $pid)"
		n=$((n + 1))
	done < <(scan_injected)
	cancel_ttl
	log "reverted ${n} netns"

	local left
	left=$(scan_injected)
	if [[ -n $left ]]; then
		log "ERROR gray-failure rules still present after revert:"
		log "$left"
		exit 1
	fi
}

case ${1:-} in
apply) shift; cmd_apply "$@" ;;
status) cmd_status ;;
revert) cmd_revert ;;
*) echo "usage: $0 apply|status|revert [flags]" >&2; exit 2 ;;
esac
