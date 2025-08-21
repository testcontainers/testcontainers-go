package nebulagraph

import (
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

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
	graphdPortTCP   = "9669/tcp"
	metadPortTCP    = "9559/tcp"
	storagedPortTCP = "9779/tcp"

	graphdPortHTTP   = "19669"
	metadPortHTTP    = "19559"
	storagedPortHTTP = "19779"
)

func defaultGrapdContainerRequest(networkName string) testcontainers.GenericContainerRequest {
	return testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Name:         defaultGraphdContainerName,
			Image:        defaultGraphdImage,
			ExposedPorts: []string{graphdPortTCP, "19669/tcp"},
			Cmd: []string{
				"--meta_server_addrs=metad0:9559",
				"--port=9669",
				"--local_ip=graphd",
				"--ws_ip=graphd",
				"--ws_http_port=19669",
				"--logtostderr=true",
				"--redirect_stdout=false",
				"--v=0",
				"--minloglevel=0",
			},
			Env:            map[string]string{"USER": "root"},
			WaitingFor:     wait.ForHTTP("/status").WithPort("19669/tcp").WithStartupTimeout(30 * time.Second),
			Networks:       []string{networkName},
			NetworkAliases: map[string][]string{networkName: {"graphd"}},
		},
		Started: true,
	}
}

func defaultMetadContainerRequest(networkName string) testcontainers.GenericContainerRequest {
	metadReq := testcontainers.ContainerRequest{
		Name:         defaultMetadContainerName,
		Image:        defaultMetadImage,
		ExposedPorts: []string{metadPortTCP, "19559/tcp"},
		Cmd: []string{
			"--meta_server_addrs=metad0:9559",
			"--local_ip=metad0",
			"--ws_ip=metad0",
			"--port=9559",
			"--ws_http_port=19559",
			"--data_path=/data/meta",
			"--logtostderr=true",
			"--redirect_stdout=false",
			"--v=0",
			"--minloglevel=0",
		},
		Env:            map[string]string{"USER": "root"},
		WaitingFor:     wait.ForHTTP("/status").WithPort("19559/tcp").WithStartupTimeout(2 * time.Minute),
		Networks:       []string{networkName},
		NetworkAliases: map[string][]string{ /*networkName: {"metad0"}*/ },
	}
	return testcontainers.GenericContainerRequest{
		ContainerRequest: metadReq,
		Started:          true,
	}
}

func defaultStoragedContainerRequest(networkName string) testcontainers.GenericContainerRequest {
	return testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Name:         defaultStoragedContainerName,
			Image:        defaultStoragedImage,
			ExposedPorts: []string{storagedPortTCP, "19779/tcp"},
			Cmd: []string{
				"--meta_server_addrs=metad0:9559",
				"--local_ip=storaged0",
				"--ws_ip=storaged0",
				"--port=9779",
				"--ws_http_port=19779",
				"--data_path=/data/storage",
				"--logtostderr=true",
				"--redirect_stdout=false",
				"--v=0",
				"--minloglevel=0",
			},
			Env:            map[string]string{"USER": "root"},
			WaitingFor:     wait.ForLog("localhost = \"storaged0\":9779").WithStartupTimeout(30 * time.Second),
			Networks:       []string{networkName},
			NetworkAliases: map[string][]string{networkName: {"storaged0"}},
		},
		Started: true,
	}
}

func defaultActivatorContainerRequest(networkName string) testcontainers.GenericContainerRequest {
	return testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:      defaultNebulaConsoleImage,
			Entrypoint: []string{},
			Cmd: []string{
				"sh", "-c",
				`for i in $(seq 1 ${ACTIVATOR_RETRY}); do
					echo "nebula" | nebula-console -addr graphd -port 9669 -u root -e 'ADD HOSTS "storaged0":9779' 1>/dev/null 2>/dev/null
					if [ $? -eq 0 ]; then
						echo "✔️ Storage activated successfully."
						break
					else
						output=$(echo "nebula" | nebula-console -addr graphd -port 9669 -u root -e 'ADD HOSTS "storaged0":9779' 2>&1)
						if echo "$output" | grep -q "Existed"; then
							echo "✔️ Storage already activated, Exiting..."
							break
						fi
					fi
					if [ $i -lt ${ACTIVATOR_RETRY} ]; then
						echo "⏳ Attempting to activate storaged, attempt $i/${ACTIVATOR_RETRY}... It's normal to take some attempts before storaged is ready. Please wait."
					else
						echo "❌ Failed to activate storaged after ${ACTIVATOR_RETRY} attempts. Please check MetaD, StorageD logs."
						echo "ℹ️ Error during storage activation:"
						echo "=============================================================="
						echo "$output"
						echo "=============================================================="
						exit 1
					fi
					sleep 5
				done && tail -f /dev/null`,
			},
			Networks:       []string{networkName},
			NetworkAliases: map[string][]string{networkName: {"activator"}},
			WaitingFor:     wait.ForExit().WithExitTimeout(3 * time.Minute),
			Env: map[string]string{
				"USER":            "root",
				"ACTIVATOR_RETRY": "30",
			},
		},
		Started: false,
	}
}
