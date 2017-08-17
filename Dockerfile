# create container stage to build out package
FROM alpine:3.6

# copy the source code into the container
COPY ./ /tmp/geoipupdate/
# run all of our commands inside the project source code
WORKDIR /tmp/geoipupdate
# install the development dependencies needed to compile the project
RUN apk add --update \
        autoconf \
        automake \
        libtool \
        alpine-sdk \
        zlib-dev \
        curl-dev

# build the project
RUN ./bootstrap
RUN ./configure
RUN make
RUN make install

# create a new image that contains only the binary
FROM alpine:3.6
# install the libraries needed to run the project
RUN apk add --update libcurl zlib
# copy over the compiled project files from the build step
COPY --from=0 /usr/local/bin/geoipupdate /usr/local/bin/geoipupdate
COPY --from=0 /usr/local/etc/GeoIP.conf /usr/local/etc/GeoIP.conf
COPY --from=0 /usr/local/share/doc/geoipupdate /usr/local/share/doc/geoipupdate
COPY --from=0 /usr/local/share/GeoIP /usr/local/share/GeoIP
# use the binary as the entrypoint for the image
ENTRYPOINT ["/usr/local/bin/geoipupdate"]
