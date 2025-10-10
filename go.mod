module github.com/maxmind/geoipupdate/v7

go 1.24.0

require (
	github.com/cenkalti/backoff/v5 v5.0.3
	github.com/gofrs/flock v0.13.0
	github.com/spf13/pflag v1.0.10
	github.com/stretchr/testify v1.11.1
	golang.org/x/net v0.46.0
	golang.org/x/sync v0.17.0
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rogpeppe/go-internal v1.13.1 // indirect
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/text v0.30.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// The module version (v6) did not match the tag version in this release.
retract v7.0.0
