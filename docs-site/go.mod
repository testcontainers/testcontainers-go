// This module only exists to keep the Python tooling used to build the docs site
// (https://golang.testcontainers.org) out of the github.com/testcontainers/testcontainers-go
// module: the Go toolchain excludes any directory containing a go.mod file from the parent
// module's zip, so these files no longer land in users' module caches and vendor directories,
// where they were picked up by dependency scanners. It contains no Go code and is never released.
module github.com/testcontainers/testcontainers-go/docs-site

go 1.26.0
