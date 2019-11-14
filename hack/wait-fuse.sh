#!/usr/bin/env bash

set -eo pipefail
#set -x # For debugging

TARGET_IP="127.0.0.1"
TARGET_PORT="65534"

usage() {
cat << EOF
USAGE: $0 [-a <host>] [-p <port>]
Waiting for fuse server ready
OPTIONS:
   -h <host>            Set the target host
   -p <port>            Set the target port
EXAMPLES:
   wait-fuse.sh -h 127.0.0.1 -p 65534
EOF
}

while getopts h:p: o
do	case "$o" in
	h)      TARGET_IP=$OPTARG;;
	p)      TARGET_PORT=$OPTARG;;
	[?])	usage
		exit 1;;
	esac
done

wait_for() {
    local host=${1?ERROR: A host is reqiured}
    local port=${2?ERROR: A port is reqiured}
    local delay=${3:-2} # Default is 2 seconds
    local retry=${4:-100}
    local count=0
    while true; do
        local telnet_count=`echo "exit" | telnet $host $port | grep -v "Connection refused" | grep "Connected to" | grep -v grep | wc -l`
        if [ ${telnet_count} -eq 1 ] ; then
            sleep 2
            break
        elif [ ${count} -eq ${retry} ] ; then
            each "Waiting for Fuse server ready timeout"
            exit 1
        else
            echo "Cannot connect to Fuse server at ${host}:${port}, retry..."
        fi
        sleep ${delay}
        let count=$count+1
    done
}


wait_for ${TARGET_IP} ${TARGET_PORT} 5 60
