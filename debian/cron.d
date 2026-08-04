# Regular cron job for the geoipupdate package, used to update GeoIP databases
#
# Daily, unlike Debian's weekly job; perl supplies the spread cron cannot.
# The AccountID test is stricter than Debian's; our placeholder is uncommented.
#
# m h dom mon dow user  command
47 0    * * *   root    test -x /usr/bin/geoipupdate && /usr/bin/grep -Eq '^[[:space:]]*(AccountID|UserId)[[:space:]]+0*[1-9][0-9]*' /etc/GeoIP.conf && test ! -d /run/systemd/system && /usr/bin/perl -e 'sleep int(rand(43200))' && /usr/bin/geoipupdate
