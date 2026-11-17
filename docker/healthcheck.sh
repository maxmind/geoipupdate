#!/bin/sh

set -e

# 1 hour is added to the update frequency threshold for a run in progress.
cutoff_duration=$((GEOIPUPDATE_FREQUENCY * 60 * 60 + 3600))
current_time=$(date +%s)
cutoff_date=$(($current_time - $cutoff_duration))

log_file="/tmp/geoipupdate/.healthcheck"

# entry.sh replaces the log only after a successful run. Until the first run
# succeeds, the log does not exist and the check fails. The Dockerfile's
# --start-period gives the first run 1 hour. Docker reserves exit code 2, so a
# jq error exits with 1.
if ! jq -e --argjson cutoff "$cutoff_date" 'map(.checked_at) | max > $cutoff' "$log_file" >/dev/null; then
  echo "healthcheck: no edition was checked since $cutoff_date"
  exit 1
fi
