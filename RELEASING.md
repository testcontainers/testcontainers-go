# Releasing Testcontainers for Go

In order to create a release, we have added a shell script that performs all the tasks for you, allowing a dry-run mode for checking it before creating the release. We are going to explain how to use it in this document.

## Prerequisites

First, it's really important that you first check that the [version.go](./internal/version.go) file is up-to-date, containing the right version you want to create. That file will be used by the automation to perform the release.
Once the version file is correct in the repository:

Second, check that the git remote for the `origin` is pointing to `github.com/testcontainers/testcontainers-go`. You can check it by running:

```shell
git remote -v
```

## Prepare the release

Once the remote is properly set, please follow these steps:

- Prepare the release with the following command. First you have to position your terminal in the `cmd/devtools` directory:
```shell
cd cmd/devtools
go run main.go release --dry-run
```

- You can use the `--dry-run` flag to enable or disable the dry-run mode. By default, it's disabled.
- To prepare for a release, updating the _Testcontainers for Go_ dependency for all the modules and examples, without performing any Git operation:

This step is idempotent, so you can run it as many times as you want.

```shell
cd cmd/devtools
go run main.go release --pre-release
```

- The script will update the [mkdocs.yml](./mkdocks.yml) file, updating the `latest_version` field to the current version.
- The script will update the `go.mod` files for each Go modules and example modules under the examples and modules directories, updating the version of the testcontainers-go dependency to the recently created tag.
- The script will modify the docs for the each Go module **that was not released yet**, updating the version of _Testcontainers for Go_ where it was added to the recently created tag.

An example execution, with dry-run mode enabled:

<details>
  <summary>Sample output</summary>
  
```shell
Executing 'git checkout main'
Replacing 'latest_version: .*' with 'latest_version: v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/mkdocs.yml
Replacing 'sonar.projectVersion=.*' with 'sonar.projectVersion=v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/sonar-project.properties
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/examples/nginx/go.mod
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/examples/toxiproxy/go.mod
make [tidy-examples]
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/modules/artemis/go.mod
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/modules/azurite/go.mod
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/modules/cassandra/go.mod
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/modules/chroma/go.mod
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/modules/clickhouse/go.mod
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/modules/cockroachdb/go.mod
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/modules/compose/go.mod
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/modules/consul/go.mod
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/modules/couchbase/go.mod
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/modules/dolt/go.mod
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/modules/elasticsearch/go.mod
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/modules/gcloud/go.mod
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/modules/inbucket/go.mod
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/modules/influxdb/go.mod
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/modules/k3s/go.mod
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/modules/k6/go.mod
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/modules/kafka/go.mod
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/modules/localstack/go.mod
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/modules/mariadb/go.mod
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/modules/milvus/go.mod
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/modules/minio/go.mod
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/modules/mockserver/go.mod
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/modules/mongodb/go.mod
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/modules/mssql/go.mod
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/modules/mysql/go.mod
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/modules/nats/go.mod
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/modules/neo4j/go.mod
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/modules/ollama/go.mod
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/modules/openfga/go.mod
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/modules/openldap/go.mod
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/modules/opensearch/go.mod
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/modules/postgres/go.mod
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/modules/pulsar/go.mod
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/modules/qdrant/go.mod
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/modules/rabbitmq/go.mod
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/modules/redis/go.mod
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/modules/redpanda/go.mod
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/modules/registry/go.mod
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/modules/surrealdb/go.mod
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/modules/vault/go.mod
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/modules/vearch/go.mod
Replacing 'testcontainers-go v.*' with 'testcontainers-go v0.32.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/modules/weaviate/go.mod
make [tidy-modules]
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/bounty.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/contributing.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/contributing_docs.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/examples/index.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/examples/nginx.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/examples/toxiproxy.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/features/build_from_dockerfile.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/features/common_functional_options.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/features/configuration.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/features/creating_container.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/features/creating_networks.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/features/docker_auth.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/features/docker_compose.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/features/files_and_mounts.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/features/follow_logs.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/features/garbage_collector.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/features/image_name_substitution.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/features/networking.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/features/override_container_command.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/features/test_session_semantics.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/features/tls.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/features/wait/exec.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/features/wait/exit.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/features/wait/health.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/features/wait/host_port.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/features/wait/http.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/features/wait/introduction.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/features/wait/log.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/features/wait/multi.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/features/wait/sql.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/getting_help.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/index.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/modules/artemis.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/modules/azurite.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/modules/cassandra.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/modules/chroma.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/modules/clickhouse.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/modules/cockroachdb.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/modules/consul.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/modules/couchbase.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/modules/dolt.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/modules/elasticsearch.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/modules/gcloud.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/modules/inbucket.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/modules/index.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/modules/influxdb.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/modules/k3s.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/modules/k6.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/modules/kafka.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/modules/localstack.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/modules/mariadb.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/modules/milvus.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/modules/minio.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/modules/mockserver.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/modules/mongodb.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/modules/mssql.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/modules/mysql.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/modules/nats.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/modules/neo4j.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/modules/ollama.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/modules/openfga.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/modules/openldap.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/modules/opensearch.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/modules/postgres.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/modules/pulsar.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/modules/qdrant.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/modules/rabbitmq.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/modules/redis.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/modules/redpanda.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/modules/registry.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/modules/surrealdb.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/modules/vault.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/modules/vearch.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/modules/weaviate.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/quickstart.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/system_requirements/ci/aws_codebuild.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/system_requirements/ci/bitbucket_pipelines.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/system_requirements/ci/circle_ci.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/system_requirements/ci/concourse_ci.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/system_requirements/ci/dind_patterns.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/system_requirements/ci/drone.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/system_requirements/ci/gitlab_ci.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/system_requirements/ci/tekton.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/system_requirements/ci/travis.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/system_requirements/docker.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/system_requirements/index.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/system_requirements/rancher.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/system_requirements/using_colima.md
Replacing 'Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>' with 'Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v0.32.0\"><span class=\"tc-version\">:material-tag: v0.32.0</span></a>' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/docs/system_requirements/using_podman.md
```
</details>

## Performing a release

Once you are satisfied with the modified files in the git state:

- Run the [release](./cmd/devtools/cmd/release/release.go) command script to create the release in dry-run mode.
- You can use the `--dry-run` flag to enable or disable the dry-run mode. **By default, it's disabled**.

```shell
cd cmd/devtools
go run main.go release
```

- You can define the bump type, using the `--bump-type` flag. The default value is `minor`, but you can also use `major`, `patch` or `prerel` (the script will fail if the value is not one of these four). That value represents the next development version after the release.

```shell
cd cmd/devtools
go run main.go release --bump-type=major
```

- The script will commit the current state of the git repository, if the `--dry-run` flag is set to `false`. The modified files are the ones modified by the `pre-release` stage.
- The script will create a git tag with the current value of the [version.go](./internal/version.go) file, starting with `v`: e.g. `v0.18.0`, for the following Go modules:
    - the root module, representing the Testcontainers for Go library.
    - all the Go modules living in both the `examples` and `modules` directory. The git tag value for these Go modules will be created using this name convention:

             "${directory}/${module_name}/${version}", e.g. "examples/mysql/v0.18.0", "modules/compose/v0.18.0"

- The script will update the [version.go](./internal/version.go) file, setting the next development version to the value defined in the `--bump-type` flag. For example, if the current version is `v0.18.0`, and a `minor` bump is used, the script will update the [version.go](./internal/version.go) file with the next development version `v0.19.0`.
- The script will create a commit with the modified version file in the **main** branch if the `--dry-run` flag is set to `false`. This represents the new development version after the release. The commit message is `chore: prepare for next minor development cycle (0.34.0)`.
- The script will push the main branch including the tags to the upstream repository, https://github.com/testcontainers/testcontainers-go, if the `--dry-run` flag is set to `false`.
- Finally, the script will trigger the Golang proxy to update the modules in https://proxy.golang.org/, if the `--dry-run` flag is set to `false`.

An example execution, with dry-run mode enabled:

<details>
  <summary>Sample output</summary>

```shell
cd cmd/devtools
go run main.go release --dry-run
Current version: 0.32.0
2024/07/01 15:53:57 github.com/testcontainers/testcontainers-go - Connected to docker: 
  Server Version: 78+testcontainerscloud (via Testcontainers Desktop 1.15.0)
  API Version: 1.43
  Operating System: Ubuntu 20.04 LTS
  Total Memory: 7407 MB
  Resolved Docker Host: tcp://127.0.0.1:57590
  Resolved Docker Socket Path: /var/run/docker.sock
  Test SessionID: c9dbaf7a4380ffdfc5c31e2f9168ce663f988ec7d7be1c34edfdc509c13515ba
  Test ProcessID: ccc98d96-5afc-4bee-af82-0586ef1133b4
2024/07/01 15:53:57 🐳 Creating container for image testcontainers/ryuk:0.7.0
2024/07/01 15:53:57 ✅ Container created: 8395b2ac2253
2024/07/01 15:53:57 🐳 Starting container: 8395b2ac2253
2024/07/01 15:53:58 ✅ Container started: 8395b2ac2253
2024/07/01 15:53:58 ⏳ Waiting for container id 8395b2ac2253 image: testcontainers/ryuk:0.7.0. Waiting for: &{Port:8080/tcp timeout:<nil> PollInterval:100ms}
2024/07/01 15:53:58 🔔 Container is ready: 8395b2ac2253
2024/07/01 15:53:58 Failed to get image auth for docker.io. Setting empty credentials for the image: docker.io/mdelapenya/semver-tool:3.4.0. Error is:credentials not found in native keychain
2024/07/01 15:54:00 🐳 Creating container for image docker.io/mdelapenya/semver-tool:3.4.0
2024/07/01 15:54:00 ✅ Container created: 6477461bd852
2024/07/01 15:54:00 🐳 Starting container: 6477461bd852
2024/07/01 15:54:00 ✅ Container started: 6477461bd852
2024/07/01 15:54:00 ⏳ Waiting for container id 6477461bd852 image: docker.io/mdelapenya/semver-tool:3.4.0. Waiting for: &{timeout:<nil> PollInterval:100ms}
2024/07/01 15:54:01 🔔 Container is ready: 6477461bd852
2024/07/01 15:54:01 🐳 Terminating container: 6477461bd852
2024/07/01 15:54:01 🚫 Container terminated: 6477461bd852
Next version: 0.33.0
Executing 'git add mkdocs.yml sonar-project.properties'
Executing 'git add 'docs/*.md''
Executing 'git add 'docs/**/*.md''
Executing 'git add 'examples/**/go.*''
Executing 'git add 'modules/**/go.*''
Executing 'git commit -m 'chore: use new version (v0.32.0) in modules and examples''
Executing 'git tag -d v0.32.0'
Executing 'git tag v0.32.0'
Executing 'git tag -d examples/nginx/v0.32.0'
Executing 'git tag examples/nginx/v0.32.0'
Executing 'git tag -d examples/toxiproxy/v0.32.0'
Executing 'git tag examples/toxiproxy/v0.32.0'
Executing 'git tag -d modules/artemis/v0.32.0'
Executing 'git tag modules/artemis/v0.32.0'
Executing 'git tag -d modules/azurite/v0.32.0'
Executing 'git tag modules/azurite/v0.32.0'
Executing 'git tag -d modules/cassandra/v0.32.0'
Executing 'git tag modules/cassandra/v0.32.0'
Executing 'git tag -d modules/chroma/v0.32.0'
Executing 'git tag modules/chroma/v0.32.0'
Executing 'git tag -d modules/clickhouse/v0.32.0'
Executing 'git tag modules/clickhouse/v0.32.0'
Executing 'git tag -d modules/cockroachdb/v0.32.0'
Executing 'git tag modules/cockroachdb/v0.32.0'
Executing 'git tag -d modules/compose/v0.32.0'
Executing 'git tag modules/compose/v0.32.0'
Executing 'git tag -d modules/consul/v0.32.0'
Executing 'git tag modules/consul/v0.32.0'
Executing 'git tag -d modules/couchbase/v0.32.0'
Executing 'git tag modules/couchbase/v0.32.0'
Executing 'git tag -d modules/dolt/v0.32.0'
Executing 'git tag modules/dolt/v0.32.0'
Executing 'git tag -d modules/elasticsearch/v0.32.0'
Executing 'git tag modules/elasticsearch/v0.32.0'
Executing 'git tag -d modules/gcloud/v0.32.0'
Executing 'git tag modules/gcloud/v0.32.0'
Executing 'git tag -d modules/inbucket/v0.32.0'
Executing 'git tag modules/inbucket/v0.32.0'
Executing 'git tag -d modules/influxdb/v0.32.0'
Executing 'git tag modules/influxdb/v0.32.0'
Executing 'git tag -d modules/k3s/v0.32.0'
Executing 'git tag modules/k3s/v0.32.0'
Executing 'git tag -d modules/k6/v0.32.0'
Executing 'git tag modules/k6/v0.32.0'
Executing 'git tag -d modules/kafka/v0.32.0'
Executing 'git tag modules/kafka/v0.32.0'
Executing 'git tag -d modules/localstack/v0.32.0'
Executing 'git tag modules/localstack/v0.32.0'
Executing 'git tag -d modules/mariadb/v0.32.0'
Executing 'git tag modules/mariadb/v0.32.0'
Executing 'git tag -d modules/milvus/v0.32.0'
Executing 'git tag modules/milvus/v0.32.0'
Executing 'git tag -d modules/minio/v0.32.0'
Executing 'git tag modules/minio/v0.32.0'
Executing 'git tag -d modules/mockserver/v0.32.0'
Executing 'git tag modules/mockserver/v0.32.0'
Executing 'git tag -d modules/mongodb/v0.32.0'
Executing 'git tag modules/mongodb/v0.32.0'
Executing 'git tag -d modules/mssql/v0.32.0'
Executing 'git tag modules/mssql/v0.32.0'
Executing 'git tag -d modules/mysql/v0.32.0'
Executing 'git tag modules/mysql/v0.32.0'
Executing 'git tag -d modules/nats/v0.32.0'
Executing 'git tag modules/nats/v0.32.0'
Executing 'git tag -d modules/neo4j/v0.32.0'
Executing 'git tag modules/neo4j/v0.32.0'
Executing 'git tag -d modules/ollama/v0.32.0'
Executing 'git tag modules/ollama/v0.32.0'
Executing 'git tag -d modules/openfga/v0.32.0'
Executing 'git tag modules/openfga/v0.32.0'
Executing 'git tag -d modules/openldap/v0.32.0'
Executing 'git tag modules/openldap/v0.32.0'
Executing 'git tag -d modules/opensearch/v0.32.0'
Executing 'git tag modules/opensearch/v0.32.0'
Executing 'git tag -d modules/postgres/v0.32.0'
Executing 'git tag modules/postgres/v0.32.0'
Executing 'git tag -d modules/pulsar/v0.32.0'
Executing 'git tag modules/pulsar/v0.32.0'
Executing 'git tag -d modules/qdrant/v0.32.0'
Executing 'git tag modules/qdrant/v0.32.0'
Executing 'git tag -d modules/rabbitmq/v0.32.0'
Executing 'git tag modules/rabbitmq/v0.32.0'
Executing 'git tag -d modules/redis/v0.32.0'
Executing 'git tag modules/redis/v0.32.0'
Executing 'git tag -d modules/redpanda/v0.32.0'
Executing 'git tag modules/redpanda/v0.32.0'
Executing 'git tag -d modules/registry/v0.32.0'
Executing 'git tag modules/registry/v0.32.0'
Executing 'git tag -d modules/surrealdb/v0.32.0'
Executing 'git tag modules/surrealdb/v0.32.0'
Executing 'git tag -d modules/vault/v0.32.0'
Executing 'git tag modules/vault/v0.32.0'
Executing 'git tag -d modules/vearch/v0.32.0'
Executing 'git tag modules/vearch/v0.32.0'
Executing 'git tag -d modules/weaviate/v0.32.0'
Executing 'git tag modules/weaviate/v0.32.0'
Producing a minor bump of the version, from 0.32.0 to 0.33.0
Replacing '0.32.0' with '0.33.0' in file /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/internal/version.go
Executing 'git add /Users/mdelapenya/sourcecode/src/github.com/testcontainers/testcontainers-go/internal/version.go'
Executing 'git commit -m 'chore: prepare for next minor development version cycle (0.33.0)''
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/examples/nginx/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/examples/toxiproxy/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/artemis/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/azurite/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/cassandra/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/chroma/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/clickhouse/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/cockroachdb/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/compose/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/consul/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/couchbase/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/dolt/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/elasticsearch/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/gcloud/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/inbucket/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/influxdb/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/k3s/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/k6/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/kafka/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/localstack/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/mariadb/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/milvus/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/minio/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/mockserver/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/mongodb/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/mssql/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/mysql/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/nats/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/neo4j/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/ollama/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/openfga/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/openldap/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/opensearch/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/postgres/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/pulsar/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/qdrant/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/rabbitmq/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/redis/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/redpanda/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/registry/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/surrealdb/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/vault/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/vearch/@v/v0.32.0.info
Hitting the Golang proxy: https://proxy.golang.org/github.com/testcontainers/testcontainers-go/modules/weaviate/@v/v0.32.0.info
```
</details>

Right after that, you have to:
- Verify that the commits are in the upstream repository, otherwise, update it with the current state of the main branch.
