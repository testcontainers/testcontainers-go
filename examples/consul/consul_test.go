package consul

import (
	"bytes"
	"context"
	"testing"

	"github.com/hashicorp/consul/api"
)

func TestConsul(t *testing.T) {
	ctx := context.Background()

	container, err := startContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// perform assertions
	cfg := api.DefaultConfig()
	cfg.Address = container.endpoint
	client, err := api.NewClient(cfg)
	if nil != err {
		t.Fatal(err)
	}
	bs := []byte("apple")
	_, err = client.KV().Put(&api.KVPair{
		Key:   "fruit",
		Value: bs,
	}, nil)
	if nil != err {
		t.Fatal(err)
	}
	pair, _, err := client.KV().Get("fruit", nil)
	if err != nil {
		t.Fatal(err)
	}
	if pair.Key != "fruit" || !bytes.Equal(pair.Value, []byte("apple")) {
		t.Errorf("get KV: %v %s,expect them to be: 'fruit' and 'apple'\n", pair.Key, pair.Value)
	}
}
