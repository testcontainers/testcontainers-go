package milvus_test

import (
	"context"
	"testing"

	"github.com/milvus-io/milvus-sdk-go/v2/client"

	"github.com/testcontainers/testcontainers-go/modules/milvus"
)

func TestMilvus(t *testing.T) {
	ctx := context.Background()

	container, err := milvus.Run(ctx, "milvusdb/milvus:v2.3.9")
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	t.Run("Connect to Milvus with gRPC", func(tt *testing.T) {
		// connectionString {
		connectionStr, err := container.ConnectionString(ctx)
		// }
		if err != nil {
			tt.Fatal(err)
		}

		milvusClient, err := client.NewGrpcClient(context.Background(), connectionStr)
		if err != nil {
			tt.Fatal("failed to connect to Milvus:", err.Error())
		}
		defer milvusClient.Close()

		v, err := milvusClient.GetVersion(ctx)
		if err != nil {
			tt.Fatal("failed to get version:", err.Error())
		}

		tt.Logf("Milvus version: %s", v)
	})
}
