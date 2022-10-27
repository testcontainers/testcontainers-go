package testcontainers

import (
	"context"

	"github.com/docker/docker/client"
)

func logDockerServerInfo(ctx context.Context, client client.APIClient, logger Logging) {
	infoMessage := `%v - Connected to docker: 
  Server Version: %v
  API Version: %v
  Operating System: %v
  Total Memory: %v MB
`

	info, err := client.Info(ctx)
	if err != nil {
		logger.Printf("failed getting information about docker server: %s", err)
	}

	logger.Printf(infoMessage, packagePath,
		info.ServerVersion, client.ClientVersion(),
		info.OperatingSystem, info.MemTotal/1024/1024)
}
