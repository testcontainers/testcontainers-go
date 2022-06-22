module github.com/testcontainers/testcontainers-go

go 1.16

require (
	github.com/cenkalti/backoff/v4 v4.1.2
	github.com/containerd/containerd v1.6.2
	github.com/docker/cli v20.10.12+incompatible
	github.com/docker/compose/v2 v2.6.0
	github.com/docker/docker v20.10.11+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/docker/go-units v0.4.0
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/go-sql-driver/mysql v1.6.0
	github.com/google/uuid v1.3.0
	github.com/magiconair/properties v1.8.6
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6
	github.com/opencontainers/image-spec v1.0.2
	github.com/stretchr/testify v1.7.0
	github.com/tonistiigi/go-rosetta v0.0.0-20201102221648-9ba854641817 // indirect
	golang.org/x/sys v0.0.0-20220412211240-33da011f77ad
	gopkg.in/yaml.v3 v3.0.0
	gotest.tools/gotestsum v1.7.0
	gotest.tools/v3 v3.2.0
)

replace (
	github.com/docker/cli => github.com/docker/cli v20.10.3-0.20220309205733-2b52f62e9627+incompatible
	github.com/docker/docker => github.com/docker/docker v20.10.3-0.20220309172631-83b51522df43+incompatible
)
