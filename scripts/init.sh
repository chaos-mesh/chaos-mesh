#!/bin/sh

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

set -eo pipefail

DATADIR=""
FUSEDIR=""
SCRIPTSDIR="/tmp/scripts"

usage() {
cat << EOF
USAGE: $0 [-d data directory] [-f fuse directory]
Used to do some preparation
OPTIONS:
   -h                      Show this message
   -d <data directory>     Data directory of the application
   -f <fuse directory>     Data directory of the fuse original directory
   -s <scripts directory>  Scripts directory
EXAMPLES:
   init.sh -d /var/lib/tikv/data -f /var/lib/tikv/fuse-data
EOF
}

while getopts h:d:f: o
do	case "$o" in
	h)      usage
            exit 1;;
	d)      DATADIR=$OPTARG;;
	f)      FUSEDIR=$OPTARG;;
	[?])	usage
		exit 1;;
	esac
done

if [ ! "$DATADIR" ];then
   echo "data directory is required"
   exit 1
fi

if [ ! "$FUSEDIR" ];then
   echo "fuse directory is required"
   exit 1
fi

mkdir_dir() {
  echo "mkdir -p $1"
  mkdir -p $1

  echo "mkdir -p $2"
  mkdir -p $2
}

copy_scripts() {
  echo "mkdir -p ${1}"
  mkdir -p ${1}

  echo "cp -R /scripts/* ${1}/"
  cp -R /scripts/* ${1}/
}

copy_scripts ${SCRIPTSDIR}

mkdir_dir ${DATADIR} ${FUSEDIR}
