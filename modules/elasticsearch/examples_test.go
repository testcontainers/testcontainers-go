package elasticsearch_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	es "github.com/elastic/go-elasticsearch/v8"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/elasticsearch"
)

func ExampleRunContainer() {
	// runElasticsearchContainer {
	ctx := context.Background()
	elasticsearchContainer, err := elasticsearch.RunContainer(ctx, testcontainers.WithImage("docker.elastic.co/elasticsearch/elasticsearch:8.9.0"))
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := elasticsearchContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()
	// }

	state, err := elasticsearchContainer.State(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRunContainer_withUsingPassword() {
	// usingPassword {
	ctx := context.Background()
	elasticsearchContainer, err := elasticsearch.RunContainer(
		ctx,
		testcontainers.WithImage("docker.elastic.co/elasticsearch/elasticsearch:7.9.2"),
		elasticsearch.WithPassword("foo"),
	)
	if err != nil {
		panic(err)
	}
	defer func() {
		err := elasticsearchContainer.Terminate(ctx)
		if err != nil {
			panic(err)
		}
	}()
	// }

	fmt.Println(strings.HasPrefix(elasticsearchContainer.Settings.Address, "http://"))
	fmt.Println(elasticsearchContainer.Settings.Password)

	// Output:
	// true
	// foo
}

func ExampleRunContainer_connectUsingElasticsearchClient() {
	// elasticsearchClient {
	ctx := context.Background()
	elasticsearchContainer, err := elasticsearch.RunContainer(
		ctx,
		testcontainers.WithImage("docker.elastic.co/elasticsearch/elasticsearch:8.9.0"),
		elasticsearch.WithPassword("foo"),
	)
	if err != nil {
		panic(err)
	}
	defer func() {
		err := elasticsearchContainer.Terminate(ctx)
		if err != nil {
			panic(err)
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
		panic(err)
	}

	resp, err := esClient.Info()
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	// }

	var esResp ElasticsearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&esResp); err != nil {
		panic(err)
	}

	fmt.Println(esResp.Tagline)
	// Output: You Know, for Search
}
