#!/usr/bin/env bash
#
# Don't run this script directly, use `yarn bootstrap` to exec it.

# Check node deps.
if [[ ! -d node_modules ]]; then
  echo "No node_modules found. Install by yarn:"

  yarn
else
  echo "Already install node deps."
fi

CHAOS_DASHBOARD_BIN=../images/chaos-dashboard/bin

# Check dashboard server.
if [[ ! -f $CHAOS_DASHBOARD_BIN/chaos-dashboard ]]; then
  echo "No chaos-dashboard binary found. Install by make IN_DOCKER=1 images/chaos-dashboard/bin/chaos-dashboard:"

  cd ..
  make IN_DOCKER=1 SWAGGER=1 images/chaos-dashboard/bin/chaos-dashboard && rm -f docs/docs.go && GO111MODULE=on go mod tidy
  cd -
else
  echo "Already build chaos-dashboard."
fi
