FROM docker.io/alpine:3

COPY geoipupdate /usr/bin/geoipupdate
COPY docker/entry.sh /usr/bin/entry.sh
VOLUME [ "/usr/share/GeoIP" ]
RUN adduser -D -h /var/lib/geoip -u 1000 geoip
USER geoip
WORKDIR /var/lib/geoip
ENTRYPOINT ["/usr/bin/entry.sh"]
