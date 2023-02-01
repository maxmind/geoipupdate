FROM docker.io/alpine:3

COPY geoipupdate /usr/bin/geoipupdate
COPY docker/entry.sh /usr/bin/entry.sh
VOLUME [ "/usr/share/GeoIP" ]
RUN adduser -D -h /var/lib/geoip -u 1000 geoip
USER geoip
WORKDIR /var/lib/geoip
ENTRYPOINT ["/usr/bin/entry.sh"]

# The health check is done by checking if the database directory is modified within the
# update period minus 1 minute. The 1 minute is a threshold for allowing slower starts.
# Without the LockFile in the database directory, this check is not going to be working
# since database files are not going to be modified when there are no updates.
HEALTHCHECK --interval=10s --timeout=10s \
    CMD test $(stat -c %Y /usr/share/GeoIP) -gt $(($(date +%s) - $GEOIPUPDATE_FREQUENCY * 60 * 61 )) \
    || exit 1;