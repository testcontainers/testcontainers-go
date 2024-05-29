package network

import (
	"context"
	"testing"
)

func TestGetGatewayIP(t *testing.T) {
	// When using docker compose with DinD mode, and using host port or http wait strategy
	// It's need to invoke GetGatewayIP for get the host
	ip, err := GetGatewayIP(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if ip == "" {
		t.Fatal("could not get gateway ip")
	}
}
