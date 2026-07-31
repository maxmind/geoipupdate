# Refresh the GeoIP databases twice a week. On systemd systems
# geoipupdate.timer does this instead; the /run/systemd/system test stops both.

SHELL=/bin/sh
PATH=/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin

0 4 * * wed,sat root test -x /usr/bin/geoipupdate && test ! -d /run/systemd/system && grep -Eq '^[[:space:]]*AccountID[[:space:]]+0*[1-9][0-9]*' /etc/GeoIP.conf && perl -e 'sleep int(rand(10800))' && /usr/bin/geoipupdate
