# Contributing

We follow the same guidelines used by testcontainers-java - please [check them
out](https://www.testcontainers.org/contributing/).

## Troubleshooting Travis

If you want to reproduce a Travis build locally, please follow this instructions to spin up a Travis build agent locally:
```shell
export BUILDID="build-testcontainers"
export INSTANCE="travisci/ci-sardonyx:packer-1564753982-0c06deb6"
docker run --name $BUILDID -w /root/go/src/github.com/testcontainers/testcontainers-go -v /Users/mdelapenya/sourcecode/src/github.com/mdelapenya/testcontainers-go:/root/go/src/github.com/testcontainers/testcontainers-go -v /var/run/docker.sock:/var/run/docker.sock -dit $INSTANCE /sbin/init
```

Once the container has been created, enter it (`docker exec -ti $BUILDID bash`) and reproduce Travis steps:

```shell
eval "$(gimme 1.11.4)"
export GO111MODULE=on
export GOPATH="/root/go"
export PATH="$GOPATH/bin:$PATH"
go get gotest.tools/gotestsum
go mod tidy
go fmt ./...
go vet ./...
gotestsum --format short-verbose ./...
```
