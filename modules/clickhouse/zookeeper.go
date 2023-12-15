package clickhouse

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const defaultZKImage = "zookeeper:3.7"

const zkPort = nat.Port("2181/tcp")

// ClickHouseContainer represents the ClickHouse container type used in the module
type ZookeeperContainer struct {
	testcontainers.Container
	exposedPort nat.Port
	ipaddr      string
}

func (c *ZookeeperContainer) IP() string {
	return c.ipaddr
}

func (c *ZookeeperContainer) Port() nat.Port {
	return c.exposedPort
}

func (c *ZookeeperContainer) ClickHouseConfig() string {
	return fmt.Sprintf(`<?xml version="1.0"?>
<clickhouse>
    <logger>
        <level>debug</level>
        <console>true</console>
        <log remove="remove"/>
        <errorlog remove="remove"/>
    </logger>

    <query_log>
        <database>system</database>
        <table>query_log</table>
    </query_log>

    <timezone>Europe/Berlin</timezone>

    <zookeeper>
        <node index="1">
            <host>%s</host>
            <port>2181</port>
        </node>
    </zookeeper>

    <remote_servers>
        <default>
            <shard>
                <replica>
                    <host>localhost</host>
                    <port>9000</port>
                </replica>
            </shard>
        </default>
    </remote_servers>
    <macros>
        <cluster>default</cluster>
        <shard>shard</shard>
        <replica>replica</replica>
    </macros>

    <distributed_ddl>
        <path>/clickhouse/task_queue/ddl</path>
    </distributed_ddl>

    <format_schema_path>/var/lib/clickhouse/format_schemas/</format_schema_path>
</clickhouse>
`, c.IP())
}

func RunZookeeper(ctx context.Context) (*ZookeeperContainer, error) {
	zkcontainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			ExposedPorts: []string{zkPort.Port()},
			Image:        defaultZKImage,
			WaitingFor:   wait.ForListeningPort(zkPort),
		},
		Started: true,
	})
	if err != nil {
		return nil, err
	}

	zkExposedPort, err := zkcontainer.MappedPort(ctx, zkPort)
	if err != nil {
		return nil, err
	}

	ipaddr, err := zkcontainer.ContainerIP(ctx)
	if err != nil {
		return nil, err
	}
	return &ZookeeperContainer{Container: zkcontainer, exposedPort: zkExposedPort, ipaddr: ipaddr}, nil
}

// WithZookeeper pass a config to connect clickhouse with zookeeper and make clickhouse as cluster
// this option is not compatible with WithYamlConfigFile / WithConfigFile options
func WithZookeeper(container *ZookeeperContainer) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		f, err := os.CreateTemp("", "clickhouse-tc-config-")
		if err != nil {
			panic(err)
		}

		defer f.Close()

		// write data to the temporary file
		data := []byte(container.ClickHouseConfig())
		if _, err := f.Write(data); err != nil {
			panic(err)
		}
		cf := testcontainers.ContainerFile{
			HostFilePath:      f.Name(),
			ContainerFilePath: "/etc/clickhouse-server/config.d/config.xml",
			FileMode:          0755,
		}
		req.Files = append(req.Files, cf)
	}
}
