package testcontainers_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
)

func ExampleNoopCustomizer() {
	// A module can read its own options from this struct while the embedded
	// NoopCustomizer satisfies ContainerCustomizer without changing the request.
	type moduleOption struct {
		testcontainers.NoopCustomizer
		Username string
	}

	option := moduleOption{Username: "testuser"}
	var customizer testcontainers.ContainerCustomizer = option
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{Image: "alpine:latest"},
	}
	err := customizer.Customize(&req)
	fmt.Println(option.Username, req.Image, err)

	// Output: testuser alpine:latest <nil>
}

func TestNoopCustomizer(t *testing.T) {
	type moduleOption struct {
		testcontainers.NoopCustomizer
	}

	for _, customizer := range []testcontainers.ContainerCustomizer{
		testcontainers.NoopCustomizer{},
		&testcontainers.NoopCustomizer{},
		moduleOption{},
		&moduleOption{},
	} {
		req := testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{Image: "alpine:latest"},
			Started:          true,
		}
		original := req
		require.NoError(t, customizer.Customize(&req))
		require.Equal(t, original, req)
		require.NoError(t, customizer.Customize(nil))
	}
}
