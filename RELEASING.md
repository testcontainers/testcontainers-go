# Releasing Testcontainers for Go

In order to create a release, we have added a shell script that performs all the tasks for you, allowing a dry-run mode for checking it before creating the release. We are going to explain how to use it in this document.

First, it's really important that you first check that the [version.go](./internal/version.go) file is up-to-date, containing the right version you want to create. That file will be used by the automation to perform the release.
Once the version file is correct in the repository:

- Run the [release.sh](./scripts/release.sh) shell script to run it in dry-run mode.
- You can use the `DRY_RUN`variable to enable or disable the dry-run mode. By default, it's enabled.
- You can use the `COMMIT` variable to enable or disable the commit creation. By default, it's disabled.
- To update the _Testcontainers for Go_ dependency for all the modules and examples, without performing any Git operation, nor creating a release:

        DRY_RUN="false" COMMIT="false" ./scripts/release.sh

- To perform a release:

        DRY_RUN="false" COMMIT="true" ./scripts/release.sh

- The script will create a git tag with the current value of the [version.go](./internal/version.go) file, starting with `v`: e.g. `v0.18.0`, for the following Go modules:
    - the root module, representing the Testcontainers for Go library.
    - all the Go modules living in both the `examples` and `modules` directory. The git tag value for these Go modules will be created using this name convention:

             "${directory}/${module_name}/${version}", e.g. "examples/mysql/v0.18.0", "modules/compose/v0.18.0"

- The script will update the [mkdocs.yml](./mkdocks.yml) file, updating the `latest_version` field to the current version.
- The script will update the [version.go](./internal/version.go) file, setting the next development version to the next **minor** version by default. For example, if the current version is `v0.18.0`, the script will update the [version.go](./internal/version.go) file with the next development version `v0.19.0`.
- You can define the bump type, using the `BUMP_TYPE` environment variable. The default value is `minor`, but you can also use `major` or `patch` (the script will fail if the value is not one of these three):

        BUMP_TYPE="major" ./scripts/release.sh

- The script will update the `go.mod` files for each Go modules and example modules under the examples and modules directories, updating the version of the testcontainers-go dependency to the recently created tag.
- The script will create a commit in the **main** branch if the `COMMIT` variable is set to `true`.
- The script will push the git the main branch including the tags to the upstream repository, https://github.com/testcontainers/testcontainers-go, if the `COMMIT` variable is set to `true`.

An example execution, with dry-run mode enabled:

```
$ ./scripts/release.sh
git tag v0.19.0
git tag examples/bigtable/v0.19.0
git tag examples/cockroachdb/v0.19.0
git tag examples/consul/v0.19.0
git tag examples/datastore/v0.19.0
git tag examples/firestore/v0.19.0
git tag examples/mongodb/v0.19.0
git tag examples/mysql/v0.19.0
git tag examples/nginx/v0.19.0
git tag examples/postgres/v0.19.0
git tag examples/pubsub/v0.19.0
git tag examples/pulsar/v0.19.0
git tag examples/redis/v0.19.0
git tag examples/spanner/v0.19.0
git tag examples/toxiproxy/v0.19.0
git tag modules/compose/v0.19.0
git tag modules/localstack/v0.19.0
git stash
git checkout main
sed "s/const Version = ".*"/const Version = "0.20.0"/g" /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/internal/version.go > /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/internal/version.go.tmp
mv /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/internal/version.go.tmp /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/internal/version.go
sed "s/latest_version: .*/latest_version: v0.19.0/g" /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/mkdocs.yml > /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/mkdocs.yml.tmp
mv /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/mkdocs.yml.tmp /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/mkdocs.yml
sed "s/testcontainers-go .*/testcontainers-go v0.19.0/g" examples/bigtable/go.mod > examples/bigtable/go.mod.tmp
mv examples/bigtable/go.mod.tmp examples/bigtable/go.mod
sed "s/testcontainers-go .*/testcontainers-go v0.19.0/g" examples/cockroachdb/go.mod > examples/cockroachdb/go.mod.tmp
mv examples/cockroachdb/go.mod.tmp examples/cockroachdb/go.mod
sed "s/testcontainers-go .*/testcontainers-go v0.19.0/g" examples/consul/go.mod > examples/consul/go.mod.tmp
mv examples/consul/go.mod.tmp examples/consul/go.mod
sed "s/testcontainers-go .*/testcontainers-go v0.19.0/g" examples/datastore/go.mod > examples/datastore/go.mod.tmp
mv examples/datastore/go.mod.tmp examples/datastore/go.mod
sed "s/testcontainers-go .*/testcontainers-go v0.19.0/g" examples/firestore/go.mod > examples/firestore/go.mod.tmp
mv examples/firestore/go.mod.tmp examples/firestore/go.mod
sed "s/testcontainers-go .*/testcontainers-go v0.19.0/g" examples/mongodb/go.mod > examples/mongodb/go.mod.tmp
mv examples/mongodb/go.mod.tmp examples/mongodb/go.mod
sed "s/testcontainers-go .*/testcontainers-go v0.19.0/g" examples/mysql/go.mod > examples/mysql/go.mod.tmp
mv examples/mysql/go.mod.tmp examples/mysql/go.mod
sed "s/testcontainers-go .*/testcontainers-go v0.19.0/g" examples/nginx/go.mod > examples/nginx/go.mod.tmp
mv examples/nginx/go.mod.tmp examples/nginx/go.mod
sed "s/testcontainers-go .*/testcontainers-go v0.19.0/g" examples/postgres/go.mod > examples/postgres/go.mod.tmp
mv examples/postgres/go.mod.tmp examples/postgres/go.mod
sed "s/testcontainers-go .*/testcontainers-go v0.19.0/g" examples/pubsub/go.mod > examples/pubsub/go.mod.tmp
mv examples/pubsub/go.mod.tmp examples/pubsub/go.mod
sed "s/testcontainers-go .*/testcontainers-go v0.19.0/g" examples/pulsar/go.mod > examples/pulsar/go.mod.tmp
mv examples/pulsar/go.mod.tmp examples/pulsar/go.mod
sed "s/testcontainers-go .*/testcontainers-go v0.19.0/g" examples/redis/go.mod > examples/redis/go.mod.tmp
mv examples/redis/go.mod.tmp examples/redis/go.mod
sed "s/testcontainers-go .*/testcontainers-go v0.19.0/g" examples/spanner/go.mod > examples/spanner/go.mod.tmp
mv examples/spanner/go.mod.tmp examples/spanner/go.mod
sed "s/testcontainers-go .*/testcontainers-go v0.19.0/g" examples/toxiproxy/go.mod > examples/toxiproxy/go.mod.tmp
mv examples/toxiproxy/go.mod.tmp examples/toxiproxy/go.mod
go mod tidy
go mod tidy
go mod tidy
go mod tidy
go mod tidy
go mod tidy
go mod tidy
go mod tidy
go mod tidy
go mod tidy
go mod tidy
go mod tidy
go mod tidy
go mod tidy
go mod tidy
sed "s/testcontainers-go .*/testcontainers-go v0.19.0/g" modules/compose/go.mod > modules/compose/go.mod.tmp
mv modules/compose/go.mod.tmp modules/compose/go.mod
sed "s/testcontainers-go .*/testcontainers-go v0.19.0/g" modules/localstack/go.mod > modules/localstack/go.mod.tmp
mv modules/localstack/go.mod.tmp modules/localstack/go.mod
go mod tidy
go mod tidy
git add /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/internal/version.go
git add /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/mkdocs.yml
git add examples/**/go.mod
git add modules/**/go.mod
git commit -m chore: prepare for next minor development cycle (0.20.0)
git push origin main --tags
git stash pop
curl https://proxy.golang.org/github.com/testcontainers/testcontainers-go/@v/v0.19.0
curl https://proxy.golang.org/github.com/testcontainers/testcontainers-go/examples/bigtable/@v/v0.19.0
curl https://proxy.golang.org/github.com/testcontainers/testcontainers-go/examples/cockroachdb/@v/v0.19.0
curl https://proxy.golang.org/github.com/testcontainers/testcontainers-go/examples/consul/@v/v0.19.0
curl https://proxy.golang.org/github.com/testcontainers/testcontainers-go/examples/datastore/@v/v0.19.0
curl https://proxy.golang.org/github.com/testcontainers/testcontainers-go/examples/firestore/@v/v0.19.0
curl https://proxy.golang.org/github.com/testcontainers/testcontainers-go/examples/mongodb/@v/v0.19.0
curl https://proxy.golang.org/github.com/testcontainers/testcontainers-go/examples/mysql/@v/v0.19.0
curl https://proxy.golang.org/github.com/testcontainers/testcontainers-go/examples/nginx/@v/v0.19.0
curl https://proxy.golang.org/github.com/testcontainers/testcontainers-go/examples/postgres/@v/v0.19.0
curl https://proxy.golang.org/github.com/testcontainers/testcontainers-go/examples/pubsub/@v/v0.19.0
curl https://proxy.golang.org/github.com/testcontainers/testcontainers-go/examples/pulsar/@v/v0.19.0
curl https://proxy.golang.org/github.com/testcontainers/testcontainers-go/examples/redis/@v/v0.19.0
curl https://proxy.golang.org/github.com/testcontainers/testcontainers-go/examples/spanner/@v/v0.19.0
curl https://proxy.golang.org/github.com/testcontainers/testcontainers-go/examples/toxiproxy/@v/v0.19.0
curl https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/compose/@v/v0.19.0
curl https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/localstack/@v/v0.19.0
```

Right after that, you have to:
- Verify that the commits are in the upstream repository, otherwise, update it with the current state of the main branch.
