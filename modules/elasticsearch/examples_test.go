package elasticsearch_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	es "github.com/elastic/go-elasticsearch/v8"

	"github.com/testcontainers/testcontainers-go/modules/elasticsearch"
)

func ExampleRun() {
	// runElasticsearchContainer {
	ctx := context.Background()
	elasticsearchContainer, err := elasticsearch.Run(ctx, "docker.elastic.co/elasticsearch/elasticsearch:8.9.0")
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}
	defer func() {
		if err := elasticsearchContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	// }

	state, err := elasticsearchContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
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
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}
	defer func() {
		err := elasticsearchContainer.Terminate(ctx)
		if err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	// }

	fmt.Println(strings.HasPrefix(elasticsearchContainer.Settings.Address, "http://"))
	fmt.Println(elasticsearchContainer.Settings.Password)

	// Output:
	// true
	// foo
}

func ExampleRun_connectUsingElasticsearchClient() {
	// elasticsearchClient {
	ctx := context.Background()
	elasticsearchContainer, err := elasticsearch.Run(
		ctx,
		"docker.elastic.co/elasticsearch/elasticsearch:8.9.0",
		elasticsearch.WithPassword("foo"),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}
	defer func() {
		err := elasticsearchContainer.Terminate(ctx)
		if err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()

	cfg := es.Config{
		Addresses: []string{
			elasticsearchContainer.Settings.Address,
		},
		Username: "elastic",
		Password: elasticsearchContainer.Settings.Password,
		CACert:   elasticsearchContainer.Settings.CACert,
	}

	esClient, err := es.NewClient(cfg)
	if err != nil {
		log.Fatalf("error creating the client: %s", err) // nolint:gocritic
	}

	resp, err := esClient.Info()
	if err != nil {
		log.Fatalf("error getting response: %s", err)
	}
	defer resp.Body.Close()
	// }

	var esResp ElasticsearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&esResp); err != nil {
		log.Fatalf("error decoding response: %s", err)
	}

	fmt.Println(esResp.Tagline)
	// Output: You Know, for Search
}
