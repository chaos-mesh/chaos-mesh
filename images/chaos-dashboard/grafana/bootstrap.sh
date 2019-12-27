#!/bin/bash -e

# for chaos
export CHAOS_NS=${CHAOS_NS:-chaos-ns-not-set}
export CHAOS_EVENT_DS_URL=${CHAOS_EVENT_DS_URL:-chaos-event-url-not-set}
export CHAOS_EVENT_DS_DB=${CHAOS_EVENT_DS_DB:-events}
export CHAOS_EVENT_DS_USER=${CHAOS_EVENT_DS_USER:-root}
export CHAOS_METRIC_DS_URL=${CHAOS_METRIC_DS_URL:-chaos-metric-url-not-set}

cp /home/grafana/templates/* /home/grafana/dashboards/
find /home/grafana/dashboards/ -type f -exec sed -i "s/--CHAOS_NS--/${CHAOS_NS}/g" {} \;

/run.sh
