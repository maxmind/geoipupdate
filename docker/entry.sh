#!/bin/sh

set -e

# SIGTERM-handler
term_handler() {
  if [ $pid -ne 0 ]; then
    kill -SIGTERM "$pid"
    wait "$pid"
  fi
  exit 143; # 128 + 15 -- SIGTERM
}

trap 'kill ${!}; term_handler' SIGTERM

pid=0
conf_file=/var/lib/geoipupdate/GeoIP.conf
database_dir=/usr/share/GeoIP
log_dir="/var/lib/geoipupdate"
log_file="$log_dir/.healthcheck"
flags="--output"
frequency=$((GEOIPUPDATE_FREQUENCY * 60 * 60))

if ! [ -z "$GEOIPUPDATE_CONF_FILE" ]; then
  conf_file=$GEOIPUPDATE_CONF_FILE
fi

if ! [ -z "$GEOIPUPDATE_DB_DIR" ]; then
  database_dir=$GEOIPUPDATE_DB_DIR
fi

if [ ! -z "$GEOIPUPDATE_ACCOUNT_ID_FILE" ]; then
  GEOIPUPDATE_ACCOUNT_ID=$( cat "$GEOIPUPDATE_ACCOUNT_ID_FILE" )
fi

if [ ! -z "$GEOIPUPDATE_LICENSE_KEY_FILE" ]; then
  GEOIPUPDATE_LICENSE_KEY=$( cat "$GEOIPUPDATE_LICENSE_KEY_FILE" )
fi

if [ -z "$GEOIPUPDATE_ACCOUNT_ID" ] || [ -z  "$GEOIPUPDATE_LICENSE_KEY" ] || [ -z "$GEOIPUPDATE_EDITION_IDS" ]; then
    echo "ERROR: You must set the environment variables GEOIPUPDATE_ACCOUNT_ID, GEOIPUPDATE_LICENSE_KEY, and GEOIPUPDATE_EDITION_IDS!"
    exit 1
fi

# Create an empty configuration file. All configuration is provided via
# environment variables or command line options, but geoipupdate still
# expects a configuration file to exist.
echo "# STATE: Creating configuration file at $conf_file"
touch "$conf_file"

mkdir -p $log_dir

while true; do
    echo "# STATE: Running geoipupdate"
    /usr/bin/geoipupdate -d "$database_dir" -f "$conf_file" $flags 1>$log_file
    if [ "$frequency" -eq 0 ]; then
        break
    fi

    echo "# STATE: Sleeping for $GEOIPUPDATE_FREQUENCY hours"
    sleep "$frequency" &
    pid=$!
    wait $!
done
