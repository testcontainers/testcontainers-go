package testcontainers

import "context"

// Create a network using a provider. By default it is Docker.
func ExampleNetworkProvider_CreateNetwork() {
	ctx := context.Background()
	networkName := "new-network"
	gcr := GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			Image: "nginx",
			ExposedPorts: []string{
				"80/tcp",
			},
			Networks: []string{
				networkName,
			},
		},
	}
	provider, _ := gcr.ProviderType.GetProvider()
	net, _ := provider.CreateNetwork(ctx, NetworkRequest{
		Name:           networkName,
		CheckDuplicate: true,
	})
	defer net.Remove(ctx)

	nginxC, _ := GenericContainer(ctx, gcr)
	defer nginxC.Terminate(ctx)
	nginxC.GetContainerID()
}
