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
conf_file=/etc/GeoIP.conf
database_dir=/usr/share/GeoIP
log_dir="/var/lib/geoipupdate"
log_file="$log_dir/.healthcheck"
frequency=$(($GEOIPUPDATE_FREQUENCY * 60 * 60))

mkdir -p $log_dir

if [ -n "$GEOIPUPDATE_CONF_FILE" ]; then
  conf_file=$GEOIPUPDATE_CONF_FILE
fi

if [ -n "$GEOIPUPDATE_DB_DIR" ]; then
  database_dir=$GEOIPUPDATE_DB_DIR
fi

if [ -n "$GEOIPUPDATE_ACCOUNT_ID_FILE" ]; then
  GEOIPUPDATE_ACCOUNT_ID=$( cat "$GEOIPUPDATE_ACCOUNT_ID_FILE" )
fi

if [ -n "$GEOIPUPDATE_LICENSE_KEY_FILE" ]; then
  GEOIPUPDATE_LICENSE_KEY=$( cat "$GEOIPUPDATE_LICENSE_KEY_FILE" )
fi

if [ -z "$GEOIPUPDATE_ACCOUNT_ID" ] || [ -z  "$GEOIPUPDATE_LICENSE_KEY" ] || [ -z "$GEOIPUPDATE_EDITION_IDS" ]; then
  echo "ERROR: You must set the environment variables GEOIPUPDATE_ACCOUNT_ID, GEOIPUPDATE_LICENSE_KEY, and GEOIPUPDATE_EDITION_IDS!"
  exit 1
fi

# Create configuration file
echo "# STATE: Creating configuration file at $conf_file"
cat <<EOF > "$conf_file"
AccountID $GEOIPUPDATE_ACCOUNT_ID
LicenseKey $GEOIPUPDATE_LICENSE_KEY
EditionIDs $GEOIPUPDATE_EDITION_IDS
DatabaseDirectory $database_dir
EOF

if [ -n "$GEOIPUPDATE_HOST" ]; then
  echo "Host $GEOIPUPDATE_HOST" >> "$conf_file"
fi

if [ -n "$GEOIPUPDATE_PROXY" ]; then
  echo "Proxy $GEOIPUPDATE_PROXY" >> "$conf_file"
fi

if [ -n "$GEOIPUPDATE_PROXY_USER_PASSWORD" ]; then
  echo "ProxyUserPassword $GEOIPUPDATE_PROXY_USER_PASSWORD" >> "$conf_file"
fi

if [ -n "$GEOIPUPDATE_PRESERVE_FILE_TIMES" ]; then
  echo "PreserveFileTimes $GEOIPUPDATE_PRESERVE_FILE_TIMES" >> "$conf_file"
fi

if [ -n "$GEOIPUPDATE_PARALLELISM" ]; then
  echo "Parallelism $GEOIPUPDATE_PARALLELISM" >> "$conf_file"
fi

execute() {
  if [ "$GEOIPUPDATE_VERBOSE" ]; then
    $@ 1>$log_file
  else
    $@ 1>$log_file 2>/dev/null
  fi
}

while true; do
  echo "# STATE: Running geoipupdate"
  execute "/usr/bin/geoipupdate -f $conf_file -v"
  if [ "$frequency" -eq 0 ]; then
    break
  fi

  echo "# STATE: Sleeping for $GEOIPUPDATE_FREQUENCY hours"
  sleep "$frequency" &
  pid=$!
  wait $!
done
