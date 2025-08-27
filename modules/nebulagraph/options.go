package nebulagraph

import (
	_ "embed"
	"fmt"
	"github.com/testcontainers/testcontainers-go/network"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

//go:embed activator.sh
var activatorScript string

const user = "root"

const (
	defaultNebulaConsoleImage = "vesoft/nebula-console:v3.8.0"
)

const (
	metadNetworkAlias    = "metad0"
	graphdNetworkAlias   = "graphd0"
	storagedNetworkAlias = "storaged0"
)

const (
	graphdPort   = "9669"
	metadPort    = "9559"
	storagedPort = "9779"

	graphdPortHTTP   = "19669"
	metadPortHTTP    = "19559"
	storagedPortHTTP = "19779"
)

func defaultGrapdContainerCustomizers(nw *testcontainers.DockerNetwork) []testcontainers.ContainerCustomizer {
	customizers := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts(graphdPort+"/tcp", graphdPortHTTP+"/tcp"),
		testcontainers.WithCmdArgs([]string{
			"--meta_server_addrs=" + metadNetworkAlias + ":" + metadPort,
			"--port=" + graphdPort,
			"--local_ip=" + graphdNetworkAlias,
			"--ws_ip=" + graphdNetworkAlias,
			"--ws_http_port=" + graphdPortHTTP,
			"--logtostderr=true",
			"--redirect_stdout=false",
			"--v=0",
			"--minloglevel=0",
		}...),
		testcontainers.WithEnv(map[string]string{"USER": user}),
		testcontainers.WithWaitStrategy(wait.ForHTTP("/status").WithPort(graphdPortHTTP + "/tcp").WithStartupTimeout(2 * time.Minute)),
		network.WithNetwork([]string{graphdNetworkAlias}, nw),
	}
	return customizers
}

func defaultMetadContainerCustomizers(nw *testcontainers.DockerNetwork) []testcontainers.ContainerCustomizer {
	customizers := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts(metadPort+"/tcp", metadPortHTTP+"/tcp"),
		testcontainers.WithCmdArgs([]string{
			"--meta_server_addrs=" + metadNetworkAlias + ":" + metadPort,
			"--local_ip=" + metadNetworkAlias,
			"--ws_ip=" + metadNetworkAlias,
			"--port=" + metadPort,
			"--ws_http_port=" + metadPortHTTP,
			"--data_path=/data/meta",
			"--logtostderr=true",
			"--redirect_stdout=false",
			"--v=0",
			"--minloglevel=0",
		}...),
		testcontainers.WithEnv(map[string]string{"USER": user}),
		testcontainers.WithWaitStrategy(wait.ForHTTP("/status").WithPort(metadPortHTTP + "/tcp").WithStartupTimeout(2 * time.Minute)),
		network.WithNetwork([]string{metadNetworkAlias}, nw),
	}
	return customizers
}

func defaultStoragedContainerCustomizers(nw *testcontainers.DockerNetwork) []testcontainers.ContainerCustomizer {
	customizers := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts(storagedPort+"/tcp", storagedPortHTTP+"/tcp"),
		testcontainers.WithCmdArgs([]string{
			"--meta_server_addrs=" + metadNetworkAlias + ":" + metadPort,
			"--local_ip=" + storagedNetworkAlias,
			"--ws_ip=" + storagedNetworkAlias,
			"--port=" + storagedPort,
			"--ws_http_port=" + storagedPortHTTP,
			"--data_path=/data/storage",
			"--logtostderr=true",
			"--redirect_stdout=false",
			"--v=0",
			"--minloglevel=0",
		}...),
		testcontainers.WithEnv(map[string]string{"USER": user}),
		testcontainers.WithWaitStrategy(
			wait.ForLog(fmt.Sprintf(`localhost = "%s":%s`, storagedNetworkAlias, storagedPort)).WithStartupTimeout(30 * time.Second),
		),
		network.WithNetwork([]string{storagedNetworkAlias}, nw),
	}
	return customizers
}

func defaultActivatorContainerCustomizers(nw *testcontainers.DockerNetwork) []testcontainers.ContainerCustomizer {
	customizers := []testcontainers.ContainerCustomizer{
		testcontainers.WithEntrypoint([]string{}...),
		testcontainers.WithCmd([]string{
			"sh", "-c",
			activatorScript,
		}...),
		testcontainers.WithExposedPorts(),
		testcontainers.WithEnv(map[string]string{
			"USER":            user,
			"ACTIVATOR_RETRY": "30",
		}),
		network.WithNetwork([]string{}, nw),
	}
	return customizers
}
