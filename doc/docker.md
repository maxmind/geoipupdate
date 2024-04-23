# Docker

## Image information

The image is available on [ghcr.io](https://github.com/maxmind/geoipupdate/pkgs/container/geoipupdate).
The source code is available on [GitHub](https://github.com/maxmind/geoipupdate).

## Configuring

The Docker image is configured by environment variables. The following
variables are required:

* `GEOIPUPDATE_EDITION_IDS` - List of space-separated database edition IDs.
  Edition IDs may consist of letters, digits, and dashes. For example,
  `GeoIP2-City` would download the GeoIP2 City database (`GeoIP2-City`).

One of:

* `GEOIPUPDATE_ACCOUNT_ID` - Your MaxMind account ID.
* `GEOIPUPDATE_ACCOUNT_ID_FILE` - A file containing your MaxMind account ID.

One of:

* `GEOIPUPDATE_LICENSE_KEY` - Your case-sensitive MaxMind license key.
* `GEOIPUPDATE_LICENSE_KEY_FILE` - A file containing your case-sensitive
  MaxMind license key.

The following are optional:

* `GEOIPUPDATE_FREQUENCY` - The number of hours between `geoipupdate` runs.
  If this is not set or is set to `0`, `geoipupdate` will run once and exit.
* `GEOIPUPDATE_HOST` - The host name of the server to use. The default is
  `https://updates.maxmind.com`.
* `GEOIPUPDATE_PROXY` - The proxy host name or IP address. You may optionally
  specify a port number, e.g., 127.0.0.1:8888. If no port number is specified,
  1080 will be used.
* `GEOIPUPDATE_PROXY_USER_PASSWORD` - The proxy user name and password,
  separated by a colon. For instance, `username:password`.
* `GEOIPUPDATE_PRESERVE_FILE_TIMES` - Whether to preserve modification times
  of files downloaded from the server. This option is either `0` or `1`. The
  default is `0`.
* `GEOIPUPDATE_VERBOSE` - Enable verbose mode. Prints out the steps that
  `geoipupdate` takes. Set to `1` to enable.
* `GEOIPUPDATE_CONF_FILE` - The path of a configuration file to be used by
  `geoipupdate`.
* `GEOIPUPDATE_DB_DIR` - The directory where geoipupdate will download the
  databases. The default is `/usr/share/GeoIP`.

The environment variables can be placed in a file with one per line and
passed in with the `--env-file` flag. Alternatively, you may pass them in
individually with the `-e` flag.

## Running

### docker run

Run the latest image with:

```sh
docker run --env-file <file> -v <database directory>:/usr/share/GeoIP ghcr.io/maxmind/geoipupdate
```

`<file>` should be the environment variable file with your configuration.
`<database directory>` should be the local directory that you want to download
the databases to.

### docker-compose

Run the latest image with:

```yaml
version: '3'
services:
  geoipupdate:
    container_name: geoipupdate
    image: ghcr.io/maxmind/geoipupdate
    restart: unless-stopped
    environment:
      - GEOIPUPDATE_ACCOUNT_ID=XXXXXX
      - GEOIPUPDATE_LICENSE_KEY=XXXXXXXXXXXXXXXX
      - 'GEOIPUPDATE_EDITION_IDS=GeoLite2-ASN GeoLite2-City GeoLite2-Country'
      - GEOIPUPDATE_FREQUENCY=72
    networks:
      - geoipupdate
    volumes:
      - 'geoipupdate_data:/usr/share/GeoIP'

networks:
  geoipupdate:

volumes:
  geoipupdate_data:
    driver: local
```

You may also pass your MaxMind account ID and license key as secrets, for example:

```yaml
version: '3'
services:
  geoipupdate:
    container_name: geoipupdate
    image: ghcr.io/maxmind/geoipupdate
    restart: unless-stopped
    environment:
      - 'GEOIPUPDATE_ACCOUNT_ID_FILE=/run/secrets/GEOIPUPDATE_ACCOUNT_ID'
      - 'GEOIPUPDATE_LICENSE_KEY_FILE=/run/secrets/GEOIPUPDATE_LICENSE_KEY'
      - 'GEOIPUPDATE_EDITION_IDS=GeoLite2-ASN GeoLite2-City GeoLite2-Country'
      - GEOIPUPDATE_FREQUENCY=72
    networks:
      - geoipupdate
    volumes:
      - 'geoipupdate_data:/usr/share/GeoIP'
    secrets:
      - GEOIPUPDATE_ACCOUNT_ID
      - GEOIPUPDATE_LICENSE_KEY

networks:
  geoipupdate:

volumes:
  geoipupdate_data:
    driver: local

secrets:
  GEOIPUPDATE_ACCOUNT_ID:
    file: ./secrets/GEOIPUPDATE_ACCOUNT_ID.txt
  GEOIPUPDATE_LICENSE_KEY:
    file: ./secrets/GEOIPUPDATE_LICENSE_KEY.txt
```

Note - When using docker-compose, you need to either:

* set `GEOIPUPDATE_FREQUENCY` equal to something greater than 0 or
* set `restart: on-failure`

If you don't, the container will continuously restart.
