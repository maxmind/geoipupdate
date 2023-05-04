FROM alpine:3

RUN apk update && \
    apk add jq

COPY geoipupdate /usr/bin/geoipupdate
COPY docker/entry.sh /usr/bin/entry.sh
COPY docker/healthcheck.sh /usr/bin/healthcheck.sh

ENTRYPOINT ["/usr/bin/entry.sh"]
HEALTHCHECK --interval=10s --timeout=10s CMD [ "/usr/bin/healthcheck.sh" ]
VOLUME [ "/usr/share/GeoIP" ]
