package etcd_test

import (
	"context"
	"fmt"
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/etcd"
)

func ExampleRun() {
	// runetcdContainer {
	ctx := context.Background()

	etcdContainer, err := etcd.Run(ctx, "gcr.io/etcd-development/etcd:v3.5.14")
	defer func() {
		if err := testcontainers.TerminateContainer(etcdContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := etcdContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRun_cluster() {
	ctx := context.Background()

	ctr, err := etcd.Run(ctx, "gcr.io/etcd-development/etcd:v3.5.14", etcd.WithNodes("etcd-1", "etcd-2", "etcd-3"))
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	defer func() {
		if err := testcontainers.TerminateContainer(ctr); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()

	clientEndpoints, err := ctr.ClientEndpoints(ctx)
	if err != nil {
		log.Printf("failed to get client endpoints: %s", err)
		return
	}

	// we have 3 nodes, 1 cluster node and 2 child nodes
	fmt.Println(len(clientEndpoints))

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   clientEndpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Printf("failed to create etcd client: %s", err)
		return
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	_, err = cli.Put(ctx, "sample_key", "sample_value")
	if err != nil {
		log.Printf("failed to put key: %s", err)
		return
	}

	resp, err := cli.Get(ctx, "sample_key")
	if err != nil {
		log.Printf("failed to get key: %s", err)
		return
	}

	fmt.Println(len(resp.Kvs))
	fmt.Println(string(resp.Kvs[0].Value))

	// Output:
	// 3
	// 1
	// sample_value
}
