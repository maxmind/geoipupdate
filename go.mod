module github.com/maxmind/geoipupdate/v7

go 1.21.0

toolchain go1.22.3

require (
	github.com/cenkalti/backoff/v4 v4.3.0
	github.com/gofrs/flock v0.12.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.9.0
	golang.org/x/net v0.30.0
	golang.org/x/sync v0.9.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// The module version (v6) did not match the tag version in this release.
retract v7.0.0
