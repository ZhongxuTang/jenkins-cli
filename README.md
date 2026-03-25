# jenkins-cli

A command line helper for interacting with Jenkins.

## Version metadata

`jenkins-cli version` prints the build information embedded in the binary. By default (when built locally without additional flags) it shows:

```
Version: dev
Commit: none
Build Date: unknown
```

### Setting version information at build time

Inject release metadata during build with Go's `-ldflags` option so the CLI reflects the published artifact state:

```
go build -ldflags "\ 
  -X 'github.com/lemonsoul/jenkins-cli/pkg/version.Version=v1.2.3' \
  -X 'github.com/lemonsoul/jenkins-cli/pkg/version.Commit=$(git rev-parse --short HEAD)' \
  -X 'github.com/lemonsoul/jenkins-cli/pkg/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)'" \
  -o bin/jenkins-cli
```

The same flags can be applied to `go run` or `go install` in CI pipelines. Adjust the values to match the release tag, commit SHA, and build timestamp used by your workflow.
