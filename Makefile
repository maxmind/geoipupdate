ifndef CONFFILE
CONFFILE=/usr/local/etc/GeoIP.conf
endif

ifndef DATADIR
DATADIR=/usr/local/share/GeoIP
endif

all: \
	build/geoipupdate \
	build/GeoIP.conf \
	build/GeoIP.conf.md \
	build/geoipupdate.md

build:
	mkdir -p build

build/geoipupdate: build
	(cd cmd/geoipupdate && go build -ldflags '-X main.defaultConfigFile=$(CONFFILE) -X main.defaultDatabaseDirectory=$(DATADIR)')
	cp cmd/geoipupdate/geoipupdate build

build/GeoIP.conf: build conf/GeoIP.conf.default
	sed -e 's|CONFFILE|$(CONFFILE)|g' -e 's|DATADIR|$(DATADIR)|g' conf/GeoIP.conf.default > build/GeoIP.conf

build/GeoIP.conf.md: build doc/GeoIP.conf.md
	sed -e 's|CONFFILE|$(CONFFILE)|g' -e 's|DATADIR|$(DATADIR)|g' doc/GeoIP.conf.md > build/GeoIP.conf.md

build/geoipupdate.md: build doc/geoipupdate.md
	sed -e 's|CONFFILE|$(CONFFILE)|g' -e 's|DATADIR|$(DATADIR)|g' doc/geoipupdate.md > build/geoipupdate.md

clean:
	rm -rf build
