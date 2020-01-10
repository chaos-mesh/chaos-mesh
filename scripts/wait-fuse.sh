#!/bin/sh

set -eo pipefail
#set -x # For debugging

FUSE_PID_FILE="/tmp/fuse/pid"
DELAY=5
RETRY=60

usage() {
cat << EOF
USAGE: $0 [-a <host>] [-p <port>]
Waiting for fuse server ready
OPTIONS:
   -h                   Show this message
   -f <host>            Set the target file
   -d <delay>           Set the delay time
   -r <retry>           Set the retry count
EXAMPLES:
   wait-fuse.sh -f /tmp/fuse/pid -d 5 -r 60
EOF
}

while getopts h:f:d:r: o
do	case "$o" in
	h)      usage
            exit 1;;
	f)      FUSE_PID_FILE=$OPTARG;;
	d)      DELAY=$OPTARG;;
	r)      RETRY=$OPTARG;;
	[?])	usage
		exit 1;;
	esac
done


wait_for() {
    local file=$1
    local delay=${2:-5} # Default is 2 seconds
    local retry=${3:-50}
    local coord_done=
    local count=0

    while [ -z "$coord_done" ]; do
        if [ -f "${file}" ]; then
            echo "fuse server is started"
            coord_done=1
            sleep 2
        elif [ ${count} -eq ${retry} ]; then
            echo "waiting for fuse server ready timeout"
            exit 1
        else
            echo "fuse server not running, ${file} not found, retry..."
            sleep ${delay}
            let count=$count+1
        fi
    done
}

wait_for ${FUSE_PID_FILE} ${DELAY} ${RETRY}
