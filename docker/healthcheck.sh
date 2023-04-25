#!/bin/sh

database_dir=/usr/share/GeoIP
if ! [ -z "$GEOIPUPDATE_DB_DIR" ]; then
  database_dir=$GEOIPUPDATE_DB_DIR
fi

# The health check is done by checking if the database directory is modified within the
# update period minus 1 minute. The 1 minute is a threshold for allowing slower starts.
# Without the LockFile in the database directory, this check is not going to be working
# since database files are not going to be modified when there are no updates.
cutoff_date=$(($(date +%s) - $GEOIPUPDATE_FREQUENCY * 60 * 61 ))
modified_at=$(find $datbase_dir -type f -exec stat -c "%Y" {} + | sort -nr | head -n 1)

if [[ "$modified_at" -lt "$cutoff_date" ]]; then
	exit 1
fi
