package nebulagraph

import (
	_ "embed"
	"github.com/testcontainers/testcontainers-go/network"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

//go:embed activator.sh
var activatorScript string

const (
	defaultNebulaConsoleImage = "vesoft/nebula-console:v3.8.0"

	defaultGraphdImage   = "vesoft/nebula-graphd:v3.8.0"
	defaultMetadImage    = "vesoft/nebula-metad:v3.8.0"
	defaultStoragedImage = "vesoft/nebula-storaged:v3.8.0"
)

const (
	defaultMetadContainerName    = "metad0"
	defaultGraphdContainerName   = "graphd0"
	defaultStoragedContainerName = "storaged0"
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
		testcontainers.WithName(defaultGraphdContainerName),
		testcontainers.WithExposedPorts(graphdPort+"/tcp", graphdPortHTTP+"/tcp"),
		testcontainers.WithCmdArgs([]string{
			"--meta_server_addrs=metad0:" + metadPort,
			"--port=" + graphdPort,
			"--local_ip=graphd0",
			"--ws_ip=graphd0",
			"--ws_http_port=" + graphdPortHTTP,
			"--logtostderr=true",
			"--redirect_stdout=false",
			"--v=0",
			"--minloglevel=0",
		}...),
		testcontainers.WithEnv(map[string]string{"USER": "root"}),
		testcontainers.WithWaitStrategy(wait.ForHTTP("/status").WithPort(graphdPortHTTP + "/tcp").WithStartupTimeout(2 * time.Minute)),
		network.WithNetwork([]string{"graphd0"}, nw),
	}
	return customizers
}

func defaultMetadContainerCustomizers(nw *testcontainers.DockerNetwork) []testcontainers.ContainerCustomizer {
	customizers := []testcontainers.ContainerCustomizer{
		testcontainers.WithName(defaultMetadContainerName),
		testcontainers.WithExposedPorts(metadPort+"/tcp", metadPortHTTP+"/tcp"),
		testcontainers.WithCmdArgs([]string{
			"--meta_server_addrs=metad0:" + metadPort,
			"--local_ip=metad0",
			"--ws_ip=metad0",
			"--port=" + metadPort,
			"--ws_http_port=" + metadPortHTTP,
			"--data_path=/data/meta",
			"--logtostderr=true",
			"--redirect_stdout=false",
			"--v=0",
			"--minloglevel=0",
		}...),
		testcontainers.WithEnv(map[string]string{"USER": "root"}),
		testcontainers.WithWaitStrategy(wait.ForHTTP("/status").WithPort(metadPortHTTP + "/tcp").WithStartupTimeout(2 * time.Minute)),
		network.WithNetwork([]string{"metad0"}, nw),
	}
	return customizers
}

func defaultStoragedContainerCustomizers(nw *testcontainers.DockerNetwork) []testcontainers.ContainerCustomizer {
	customizers := []testcontainers.ContainerCustomizer{
		testcontainers.WithName(defaultStoragedContainerName),
		testcontainers.WithExposedPorts(storagedPort+"/tcp", storagedPortHTTP+"/tcp"),
		testcontainers.WithCmdArgs([]string{
			"--meta_server_addrs=metad0:" + metadPort,
			"--local_ip=storaged0",
			"--ws_ip=storaged0",
			"--port=" + storagedPort,
			"--ws_http_port=" + storagedPortHTTP,
			"--data_path=/data/storage",
			"--logtostderr=true",
			"--redirect_stdout=false",
			"--v=0",
			"--minloglevel=0",
		}...),
		testcontainers.WithEnv(map[string]string{"USER": "root"}),
		testcontainers.WithWaitStrategy(wait.ForLog("localhost = \"storaged0\":" + storagedPort).WithStartupTimeout(30 * time.Second)),
		network.WithNetwork([]string{"storaged0"}, nw),
	}
	return customizers
}

func defaultActivatorContainerCustomizers(nw *testcontainers.DockerNetwork) []testcontainers.ContainerCustomizer {
	customizers := []testcontainers.ContainerCustomizer{
		testcontainers.WithName("activator"),
		testcontainers.WithEntrypoint([]string{}...),
		testcontainers.WithCmd([]string{
			"sh", "-c",
			activatorScript,
		}...),
		testcontainers.WithExposedPorts(),
		testcontainers.WithEnv(map[string]string{
			"USER":            "root",
			"ACTIVATOR_RETRY": "30",
		}),
		network.WithNetwork([]string{ /*"activator"*/ }, nw),
	}
	return customizers
}
