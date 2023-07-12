#!/bin/sh

set -e

# 2 minutes are added to the update frequency threshold to make room for slower starts.
cutoff_duration=$(($GEOIPUPDATE_FREQUENCY * 60 * 60 + 120))
current_time=$(date +%s)
cutoff_date=$(($current_time - $cutoff_duration))

log_file="/tmp/geoipupdate/.healthcheck"
editions=$(cat "$log_file" | jq -r '.[] | select(.checked_at > '$cutoff_date') | .edition_id')
checked_editions=$(echo "$editions" | wc -l)
desired_editions=$(echo "$GEOIPUPDATE_EDITION_IDS" | awk -F' ' '{print NF}')

if [ "$checked_editions" != "$desired_editions" ]; then
  echo "healtcheck editions number $checked_editions is less than the desired editions number $desired_editions"
  exit 1
fi
