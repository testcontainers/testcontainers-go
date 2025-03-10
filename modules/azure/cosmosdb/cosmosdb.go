package cosmosdb

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultPort = "8081/tcp"

	connectionStringFormat = "AccountEndpoint=%s;AccountKey=%s"

	// EmulatorKey is a known constant and specified in Azure Cosmos DB Documents.
	// This key is also used as password for emulator certificate file.
	EmulatorKey = "C2y6yDjf5/R+ob0N8A7Cgv30VRDJIWEHLM+4QDU5DE2nQ9nDuVTqobD4b8mGGyPMbIZnqyMsEcaGQy67XIw/Jw=="
)

// Container represents the Azure CosmosDB emulator container type used in the module
type Container struct {
	testcontainers.Container
}

// Run creates an instance of the Azure CosmosDB emulator container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	req := testcontainers.ContainerRequest{
		Image: img,
		Env: map[string]string{
			"AZURE_COSMOS_EMULATOR_PARTITION_COUNT":         "15",
			"AZURE_COSMOS_EMULATOR_ENABLE_DATA_PERSISTENCE": "false",
			"LOG_LEVEL": "info",
		},
		ExposedPorts: []string{defaultPort},
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, fmt.Errorf("customize: %w", err)
		}
	}

	occurrences := genericContainerReq.Env["AZURE_COSMOS_EMULATOR_PARTITION_COUNT"]
	occurrencesInt, err := strconv.Atoi(occurrences)
	if err != nil {
		return nil, fmt.Errorf("convert partition count to int: %w", err)
	}

	genericContainerReq.WaitingFor = wait.ForAll(
		wait.ForListeningPort(defaultPort).SkipInternalCheck(),
		// the log message format, for 3 partitions, is:
		// "Started 1/4 partitions"
		// "Started 2/4 partitions"
		// "Started 3/4 partitions"
		// "Started 4/4 partitions"
		// "Started"
		wait.ForLog("Started").WithStartupTimeout(time.Minute).WithOccurrence(occurrencesInt+2),
	)

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	var c *Container
	if container != nil {
		c = &Container{Container: container}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}

// ConnectionString returns the connection string for the cosmosdb emulator,
// using the following format: AccountEndpoint=</host:port>;AccountKey=<key>;
func (c *Container) ConnectionString(ctx context.Context) (string, error) {
	hostPort, err := c.PortEndpoint(ctx, defaultPort, "http")
	if err != nil {
		return "", fmt.Errorf("port endpoint: %w", err)
	}

	return fmt.Sprintf(connectionStringFormat, hostPort, EmulatorKey), nil
}

// MustConnectionString returns the connection string for the cosmosdb emulator,
// calling [Container.ConnectionString] and panicking if it returns an error.
func (c *Container) MustConnectionString(ctx context.Context) string {
	connString, err := c.ConnectionString(ctx)
	if err != nil {
		panic(err)
	}

	return connString
}

// WithPartitionCount sets the partition count for the cosmosdb emulator.
// The default is 15.
func WithPartitionCount(count int) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["AZURE_COSMOS_EMULATOR_PARTITION_COUNT"] = strconv.Itoa(count)
		return nil
	}
}

// WithLogLevel sets the log level for the cosmosdb emulator.
// The default is "info".
func WithLogLevel(level string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["LOG_LEVEL"] = level
		return nil
	}
}
