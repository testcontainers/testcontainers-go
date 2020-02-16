module github.com/erdemtoraman/testcontainers-go

replace github.com/docker/docker => github.com/docker/engine v0.0.0-20190717161051-705d9623b7c1

require (
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/docker/docker v0.7.3-0.20190506211059-b20a14b54661
	github.com/docker/go-connections v0.4.0
	github.com/gin-gonic/gin v1.5.0
	github.com/go-redis/redis v6.15.7+incompatible
	github.com/go-sql-driver/mysql v1.5.0
	github.com/pkg/errors v0.9.1
	github.com/satori/go.uuid v1.2.0
	github.com/testcontainers/testcontainers-go v0.2.0 // indirect
	gotest.tools v0.0.0-20181223230014-1083505acf35
)

go 1.13
