package gcloud

import "context"

type GCloudContainer interface {
	uri(ctx context.Context) (string, error)
}

func containerURI(ctx context.Context, container GCloudContainer) (string, error) {
	return container.uri(ctx)
}
