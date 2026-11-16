package examples_test

import (
	"context"
	"fmt"
	"log"

	influxclient2 "github.com/influxdata/influxdb-client-go/v2"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/influxdb"
)

func ExampleRun_v2() {
	// runInfluxV2Container {
	ctx := context.Background()

	username := "username"
	password := "password"
	org := "org"
	bucket := "bucket"
	token := "influxdbv2token"

	influxdbContainer, err := influxdb.Run(ctx, "influxdb:2.7.11",
		influxdb.WithV2Auth(org, bucket, username, password),
		influxdb.WithV2AdminToken(token),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(influxdbContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := influxdbContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	url, err := influxdbContainer.ConnectionUrl(ctx)
	if err != nil {
		log.Printf("failed to get host: %s", err)
		return
	}

	client := influxclient2.NewClientWithOptions(url, token, influxclient2.DefaultOptions())
	defer client.Close()

	influxBucket, err := client.BucketsAPI().FindBucketByName(ctx, bucket)
	if err != nil {
		log.Printf("failed to get bucket: %s", err)
		return
	}

	fmt.Println(influxBucket.Name)

	// Output:
	// true
	// bucket
}
