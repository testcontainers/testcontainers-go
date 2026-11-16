package elasticsearch_test

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/elasticsearch"
)

func ExampleRun() {
	// runElasticsearchContainer {
	ctx := context.Background()
	elasticsearchContainer, err := elasticsearch.Run(ctx, "docker.elastic.co/elasticsearch/elasticsearch:8.9.0")
	defer func() {
		if err := testcontainers.TerminateContainer(elasticsearchContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := elasticsearchContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRun_withUsingPassword() {
	// usingPassword {
	ctx := context.Background()
	elasticsearchContainer, err := elasticsearch.Run(
		ctx,
		"docker.elastic.co/elasticsearch/elasticsearch:7.9.2",
		elasticsearch.WithPassword("foo"),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(elasticsearchContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	fmt.Println(strings.HasPrefix(elasticsearchContainer.Settings.Address, "http://"))
	fmt.Println(elasticsearchContainer.Settings.Password)

	// Output:
	// true
	// foo
}
