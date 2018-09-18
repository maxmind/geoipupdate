ifndef CONFFILE
CONFFILE=/usr/local/etc/GeoIP.conf
endif

ifndef DATADIR
DATADIR=/usr/local/share/GeoIP
endif

all: \
	build/geoipupdate \
	build/CHANGELOG.md \
	build/LICENSE-APACHE \
	build/LICENSE-MIT \
	build/README.md \
	build/GeoIP.conf \
	build/doc/GeoIP.conf.md \
	build/doc/geoipupdate.md

build:
	mkdir -p build

build/geoipupdate: build
	(cd cmd/geoipupdate && go build -ldflags '-X main.defaultConfigFile=$(CONFFILE) -X main.defaultDatabaseDirectory=$(DATADIR)')
	cp cmd/geoipupdate/geoipupdate build

build/CHANGELOG.md: build CHANGELOG.md
	cp CHANGELOG.md build

build/LICENSE-APACHE: build LICENSE-APACHE
	cp LICENSE-APACHE build

build/LICENSE-MIT: build LICENSE-MIT
	cp LICENSE-MIT build

build/README.md: build README.md
	cp README.md build

build/GeoIP.conf: build conf/GeoIP.conf.default
	sed -e 's|CONFFILE|$(CONFFILE)|g' -e 's|DATADIR|$(DATADIR)|g' conf/GeoIP.conf.default > build/GeoIP.conf

build/doc:
	mkdir -p build/doc

build/doc/GeoIP.conf.md: build/doc doc/GeoIP.conf.md
	sed -e 's|CONFFILE|$(CONFFILE)|g' -e 's|DATADIR|$(DATADIR)|g' doc/GeoIP.conf.md > build/doc/GeoIP.conf.md

build/doc/geoipupdate.md: build/doc doc/geoipupdate.md
	sed -e 's|CONFFILE|$(CONFFILE)|g' -e 's|DATADIR|$(DATADIR)|g' doc/geoipupdate.md > build/doc/geoipupdate.md

clean:
	rm -rf build
