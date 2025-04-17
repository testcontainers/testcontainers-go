package redis

import (
	"strconv"

	"github.com/testcontainers/testcontainers-go"
)

// WithConfigFile sets the config file to be used for the redis container, and sets the command to run the redis server
// using the passed config file
func WithConfigFile(configFile string) testcontainers.CustomizeRequestOption {
	const defaultConfigFile = "/usr/local/redis.conf"

	return func(req *testcontainers.GenericContainerRequest) error {
		cf := testcontainers.ContainerFile{
			HostFilePath:      configFile,
			ContainerFilePath: defaultConfigFile,
			FileMode:          0o755,
		}
		req.Files = append(req.Files, cf)

		if len(req.Cmd) == 0 {
			req.Cmd = []string{redisServerProcess, defaultConfigFile}
			return nil
		}

		// prepend the command to run the redis server with the config file, which must be the first argument of the redis server process
		if req.Cmd[0] == redisServerProcess {
			// just insert the config file, then the rest of the args
			req.Cmd = append([]string{redisServerProcess, defaultConfigFile}, req.Cmd[1:]...)
		} else if req.Cmd[0] != redisServerProcess {
			// prepend the redis server and the config file, then the rest of the args
			req.Cmd = append([]string{redisServerProcess, defaultConfigFile}, req.Cmd...)
		}

		return nil
	}
}

// WithLogLevel sets the log level for the redis server process
// See https://redis.io/docs/reference/modules/modules-api-ref/#redismodule_log for more information.
func WithLogLevel(level LogLevel) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		processRedisServerArgs(req, []string{"--loglevel", string(level)})

		return nil
	}
}

// WithSnapshotting sets the snapshotting configuration for the redis server process. You can configure Redis to have it
// save the dataset every N seconds if there are at least M changes in the dataset.
// This method allows Redis to benefit from copy-on-write semantics.
// See https://redis.io/docs/management/persistence/#snapshotting for more information.
func WithSnapshotting(seconds int, changedKeys int) testcontainers.CustomizeRequestOption {
	if changedKeys < 1 {
		changedKeys = 1
	}
	if seconds < 1 {
		seconds = 1
	}

	return func(req *testcontainers.GenericContainerRequest) error {
		processRedisServerArgs(req, []string{"--save", strconv.Itoa(seconds), strconv.Itoa(changedKeys)})
		return nil
	}
}
