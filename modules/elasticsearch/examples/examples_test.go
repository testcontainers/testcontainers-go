package examples_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	es "github.com/elastic/go-elasticsearch/v8"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/elasticsearch"
)

func ExampleRun_connectUsingElasticsearchClient() {
	// elasticsearchClient {
	ctx := context.Background()
	elasticsearchContainer, err := elasticsearch.Run(
		ctx,
		"docker.elastic.co/elasticsearch/elasticsearch:8.9.0",
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
		log.Printf("error creating the client: %s", err)
		return
	}

	resp, err := esClient.Info()
	if err != nil {
		log.Printf("error getting response: %s", err)
		return
	}
	defer resp.Body.Close()
	// }

	var esResp struct {
		Tagline string `json:"tagline"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&esResp); err != nil {
		log.Printf("error decoding response: %s", err)
		return
	}

	fmt.Println(esResp.Tagline)
	// Output: You Know, for Search
}
