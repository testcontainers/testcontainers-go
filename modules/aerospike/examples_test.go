package aerospike_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aerospike/aerospike-client-go/v8"
	"github.com/testcontainers/testcontainers-go"
	as "github.com/testcontainers/testcontainers-go/modules/aerospike"
)

func ExampleRun() {
	// runAerospikeContainer {
	ctx := context.Background()

	aerospikedbContainer, err := as.Run(ctx, "aerospike/aerospike-server:latest")
	defer func() {
		if err := testcontainers.TerminateContainer(aerospikedbContainer.Container); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }
	state, err := aerospikedbContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRun_usingClient() {
	ctx := context.Background()

	aerospikedbContainer, err := as.Run(
		ctx, "aerospike/aerospike-server:latest",
	)
	defer func() {
		if err := testcontainers.TerminateContainer(aerospikedbContainer.Container); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	// Get the host and port
	host, err := aerospikedbContainer.Host(ctx)
	if err != nil {
		log.Printf("failed to get container host: %s", err)
		return
	}

	// Get the mapped port
	port, err := aerospikedbContainer.MappedPort(ctx, "3000/tcp")
	if err != nil {
		log.Printf("failed to get container port: %s", err)
		return
	}

	aeroHost := []*aerospike.Host{aerospike.NewHost(host, port.Int())}

	// connect to the host
	cp := aerospike.NewClientPolicy()
	cp.Timeout = 10 * time.Second

	// Create a client
	client, err := aerospike.NewClientWithPolicyAndHost(cp, aeroHost...)
	if err != nil {
		log.Printf("Failed to create aerospike client: %v", err)
		return
	}

	// Close the client
	defer client.Close()

	// Create a key
	schemaKey, err := aerospike.NewKey("test", "test", "_schema_info")
	if err != nil {
		log.Printf("Failed to create key: %v", err)
		return
	}

	version := 1
	description := "test aerospike schema info"
	nowStr := time.Now().Format(time.RFC3339)

	// Create schema record
	bins := aerospike.BinMap{
		"version":     version,
		"created_at":  nowStr,
		"updated_at":  nowStr,
		"description": description,
	}

	// Never expire the schema info
	writePolicy := aerospike.NewWritePolicy(0, 0)

	// Store in Aerospike
	err = client.Put(writePolicy, schemaKey, bins)
	if err != nil {
		log.Printf("Failed to put schema info: %v", err)
		return
	}

	// Get schema record
	record, err := client.Get(nil, schemaKey, "version", "created_at", "updated_at", "description")
	if err != nil {
		log.Printf("Failed to get schema info: %v", err)
		return
	}

	// Schema exists, check version
	existingVersion, _ := record.Bins["version"].(int)
	existingDescription, _ := record.Bins["description"].(string)

	fmt.Println(existingVersion)
	fmt.Println(existingDescription)

	// Output:
	// 1
	// test aerospike schema info
}
