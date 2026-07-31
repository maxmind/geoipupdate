# Regular cron job for the geoipupdate package, used to update GeoIP databases
#
# MaxMind typically updates their databases on Tuesdays.
#
# The AccountID test is stricter than Debian's; our placeholder is uncommented.
#
# m h dom mon dow user  command
47 6    * * 3   root    test -x /usr/bin/geoipupdate && grep -Eq '^[[:space:]]*AccountID[[:space:]]+0*[1-9][0-9]*' /etc/GeoIP.conf && test ! -d /run/systemd/system && /usr/bin/geoipupdate
