package mysql

import (
	"github.com/testcontainers/testcontainers-go"
	"path/filepath"
)

func WithUsername(username string) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		req.Env["MYSQL_USER"] = username
	}
}

func WithPassword(password string) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		req.Env["MYSQL_PASSWORD"] = password
	}
}

func WithDatabase(database string) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		req.Env["MYSQL_DATABASE"] = database
	}
}

func WithConfigFile(configFile string) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		cf := testcontainers.ContainerFile{
			HostFilePath:      configFile,
			ContainerFilePath: "/etc/mysql/conf.d/my.cnf",
			FileMode:          0755,
		}
		req.Files = append(req.Files, cf)
	}
}

func WithScripts(scripts ...string) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		var initScripts []testcontainers.ContainerFile
		for _, script := range scripts {
			cf := testcontainers.ContainerFile{
				HostFilePath:      script,
				ContainerFilePath: "/docker-entrypoint-initdb.d/" + filepath.Base(script),
				FileMode:          0755,
			}
			initScripts = append(initScripts, cf)
		}
		req.Files = append(req.Files, initScripts...)
	}
}
