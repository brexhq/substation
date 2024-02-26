# Versioning

Substation uses [Semantic Versioning 2.0](https://semver.org/). Versions are managed using Git tags and are updated by the maintainers when releases are made. The version applies to the [Go module](https://pkg.go.dev/github.com/brexhq/substation) and the components below:

- cmd/aws/*
- condition/*
- config/*
- message/*
- transform/*
- substation.go
- go.mod

Some features may be labeled as "experimental" in the documentation. These features are not subject to the same versioning guarantees as the rest of the project and may be changed or removed at any time.

## Go Versioning

Substation follows the [Go Release Policy](https://golang.org/doc/devel/release.html#policy). This means that the project will maintain compatibility with the latest two major versions of Go. For example, if the latest version of Go is 1.21, Substation will support Go 1.20 and 1.21. When Go 1.22 is released, Substation will drop support for Go 1.20 and support Go 1.21 and 1.22.

## Version Support

The maintainers will actively support the latest release of Substation with features, bug fixes, and security patches. Older versions will only receive security patches. If you are using an old version of Substation, we recommend upgrading to the latest version.
