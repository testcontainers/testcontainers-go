module github.com/testcontainers/testcontainer-go

require (
	github.com/Microsoft/go-winio v0.4.8 // indirect
	github.com/Sirupsen/logrus v1.2.0 // indirect
	github.com/docker/distribution v2.7.0-rc.0+incompatible // indirect
	github.com/docker/docker v0.7.3-0.20170522122511-6ce6ae1cd11d
	github.com/docker/go-connections v0.3.0
	github.com/docker/go-units v0.0.0-20161020213227-8a7beacffa30 // indirect
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7 // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/pkg/errors v0.0.0-20161002052512-839d9e913e06
	github.com/satori/go.uuid v1.2.0
	github.com/stevvooe/resumable v0.0.0-20180830230917-22b14a53ba50 // indirect
	golang.org/x/net v0.0.0-20180712202826-d0887baf81f4 // indirect
)

replace (
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.2.0
	github.com/docker/docker => github.com/docker/docker v0.7.3-0.20170522122511-6ce6ae1cd11d
)
