module github.com/maxmind/geoipupdate/v7

go 1.24.0

require (
	github.com/cenkalti/backoff/v5 v5.0.3
	github.com/gofrs/flock v0.13.0
	github.com/spf13/pflag v1.0.10
	github.com/stretchr/testify v1.11.1
	golang.org/x/net v0.50.0
	golang.org/x/sync v0.19.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.13.1 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/text v0.34.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// The module version (v6) did not match the tag version in this release.
retract v7.0.0
