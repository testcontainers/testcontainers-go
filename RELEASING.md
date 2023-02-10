# Releasing Testcontainers for Go

In order to create a release, we have added a shell script that performs all the tasks for you, allowing a dry-run mode for checking it before creating the release. We are going to explain how to use it in this document.

First, it's really important that you first check that the [version.go](./internal/version.go) file is up-to-date, containing the right version you want to create. That file will be used by the automation to perform the release.
Once the version file is correct in the repository:

- Run the [release.sh](./scripts/release.sh) shell script.
- You can run the script in dry-run mode setting `DRY_RUN=true` in the environment:

        DRY_RUN="true" ./scripts/release.sh

- The script will create a git tag with the current value of the [version.go](./internal/version.go) file, starting with `v`: e.g. `v0.18.0`, for the following Go modules:
    - the root module, representing the Testcontainers for Go library.
    - all the Go modules living in both the `examples` and `modules` directory. The git tag value for these Go modules will be created using this name convention:

             "${directory}/${module_name}/${version}", e.g. "examples/mysql/v0.18.0", "modules/compose/v0.18.0"

- The script will push the git tags to the upstream repository, https://github.com/testcontainers/testcontainers-go

An example execution, with dry-run mode enabled:

```
$ DRY_RUN=true ./scripts/release.sh       
git tag -d v0.18.0 | true
git tag v0.18.0
git tag -d examples/bigtable/v0.18.0 | true
git tag examples/bigtable/v0.18.0
git tag -d examples/cockroachdb/v0.18.0 | true
git tag examples/cockroachdb/v0.18.0
git tag -d examples/consul/v0.18.0 | true
git tag examples/consul/v0.18.0
git tag -d examples/datastore/v0.18.0 | true
git tag examples/datastore/v0.18.0
git tag -d examples/firestore/v0.18.0 | true
git tag examples/firestore/v0.18.0
git tag -d examples/localstack/v0.18.0 | true
git tag examples/localstack/v0.18.0
git tag -d examples/mongodb/v0.18.0 | true
git tag examples/mongodb/v0.18.0
git tag -d examples/mysql/v0.18.0 | true
git tag examples/mysql/v0.18.0
git tag -d examples/nginx/v0.18.0 | true
git tag examples/nginx/v0.18.0
git tag -d examples/postgres/v0.18.0 | true
git tag examples/postgres/v0.18.0
git tag -d examples/pubsub/v0.18.0 | true
git tag examples/pubsub/v0.18.0
git tag -d examples/pulsar/v0.18.0 | true
git tag examples/pulsar/v0.18.0
git tag -d examples/redis/v0.18.0 | true
git tag examples/redis/v0.18.0
git tag -d examples/spanner/v0.18.0 | true
git tag examples/spanner/v0.18.0
git tag -d examples/toxiproxy/v0.18.0 | true
git tag examples/toxiproxy/v0.18.0
git tag -d modules/compose/v0.18.0 | true
git tag modules/compose/v0.18.0
git push --tags
curl https://proxy.golang.org/github.com/testcontainers/testcontainers-go/@v/v0.18.0
curl https://proxy.golang.org/github.com/testcontainers/testcontainers-go/examples/bigtable/@v/v0.18.0
curl https://proxy.golang.org/github.com/testcontainers/testcontainers-go/examples/cockroachdb/@v/v0.18.0
curl https://proxy.golang.org/github.com/testcontainers/testcontainers-go/examples/consul/@v/v0.18.0
curl https://proxy.golang.org/github.com/testcontainers/testcontainers-go/examples/datastore/@v/v0.18.0
curl https://proxy.golang.org/github.com/testcontainers/testcontainers-go/examples/firestore/@v/v0.18.0
curl https://proxy.golang.org/github.com/testcontainers/testcontainers-go/examples/localstack/@v/v0.18.0
curl https://proxy.golang.org/github.com/testcontainers/testcontainers-go/examples/mongodb/@v/v0.18.0
curl https://proxy.golang.org/github.com/testcontainers/testcontainers-go/examples/mysql/@v/v0.18.0
curl https://proxy.golang.org/github.com/testcontainers/testcontainers-go/examples/nginx/@v/v0.18.0
curl https://proxy.golang.org/github.com/testcontainers/testcontainers-go/examples/postgres/@v/v0.18.0
curl https://proxy.golang.org/github.com/testcontainers/testcontainers-go/examples/pubsub/@v/v0.18.0
curl https://proxy.golang.org/github.com/testcontainers/testcontainers-go/examples/pulsar/@v/v0.18.0
curl https://proxy.golang.org/github.com/testcontainers/testcontainers-go/examples/redis/@v/v0.18.0
curl https://proxy.golang.org/github.com/testcontainers/testcontainers-go/examples/spanner/@v/v0.18.0
curl https://proxy.golang.org/github.com/testcontainers/testcontainers-go/examples/toxiproxy/@v/v0.18.0
curl https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/compose/@v/v0.18.0
```

Right after that, you have to:
- Bump the version in the [version.go](./internal/version.go) file to the next development version, following [Semantic Versioning](https://semver.org). e.g. `v0.19.0`
- Update the [mkdocs.yml](./mkdocs.yml) file with the next development version.
- Commit those files and submit a pull request, using `chore` as Github label.
