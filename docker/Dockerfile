FROM alpine:3

# The labels that vary per release (version, revision, created) are passed in by
# GoReleaser instead.
LABEL org.opencontainers.image.title="geoipupdate" \
      org.opencontainers.image.description="Program to perform automatic updates of GeoIP and GeoLite binary databases." \
      org.opencontainers.image.source="https://github.com/maxmind/geoipupdate" \
      org.opencontainers.image.url="https://github.com/maxmind/geoipupdate" \
      org.opencontainers.image.documentation="https://github.com/maxmind/geoipupdate/blob/main/README.md" \
      org.opencontainers.image.licenses="Apache-2.0 OR MIT" \
      org.opencontainers.image.vendor="MaxMind, Inc."

RUN apk update && apk add jq

COPY geoipupdate /usr/bin/geoipupdate
COPY docker/entry.sh /usr/bin/entry.sh
COPY docker/healthcheck.sh /usr/bin/healthcheck.sh

ENTRYPOINT ["/usr/bin/entry.sh"]
HEALTHCHECK --interval=10s --timeout=10s CMD [ "/usr/bin/healthcheck.sh" ]
VOLUME [ "/usr/share/GeoIP" ]
WORKDIR /var/lib/geoipupdate
