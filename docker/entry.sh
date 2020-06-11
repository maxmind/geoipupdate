#!/bin/sh

set -e

conf_file=/etc/GeoIP.conf
database_dir=/usr/share/GeoIP
flags=
frequency=$((GEOIPUPDATE_FREQUENCY * 60 * 60))
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
EOF

if [ ! -z "$GEOIPUPDATE_HOST" ]; then
    echo "Host $GEOIPUPDATE_HOST" >> "$conf_file"
fi

if [ ! -z "$GEOIPUPDATE_PROXY" ]; then
    echo "Proxy $GEOIPUPDATE_PROXY" >> "$conf_file"
fi

if [ ! -z "$GEOIPUPDATE_PROXY_USER_PASSWORD" ]; then
    echo "ProxyUserPassword $GEOIPUPDATE_PROXY_USER_PASSWORD" >> "$conf_file"
fi

if [ ! -z "$GEOIPUPDATE_PRESERVE_FILE_TIMES" ]; then
    echo "PreserveFileTimes $GEOIPUPDATE_PRESERVE_FILE_TIMES" >> "$conf_file"
fi

if [ "$GEOIPUPDATE_VERBOSE" ]; then
    flags="-v"
fi

while true; do
    echo "# STATE: Running geoipupdate"
    /usr/bin/geoipupdate -d "$database_dir" -f "$conf_file" $flags
    if [ "$frequency" -eq 0 ]; then
        break
    fi

    echo "# STATE: Sleeping for $GEOIPUPDATE_FREQUENCY hours"
    sleep "$frequency"
done
