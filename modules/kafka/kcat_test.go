package kafka_test

import (
	"context"
	"fmt"
	"io"

	"github.com/testcontainers/testcontainers-go"
)

type KcatContainer struct {
	Container testcontainers.Container
	FilePath  string
}

func createKCat(ctx context.Context, network, filepath string) (KcatContainer, error) {
	kcat, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "confluentinc/cp-kcat:7.4.1",
			Networks: []string{
				network,
			},
			Entrypoint: []string{
				"sh",
			},
			Cmd: []string{
				"-c",
				"tail -f /dev/null",
			},
		},
		Started: true,
	})

	if err != nil {
		return KcatContainer{}, fmt.Errorf("create generic container: %w", err)
	}

	return KcatContainer{Container: kcat, FilePath: filepath}, nil
}

func (kcat *KcatContainer) SaveFile(ctx context.Context, data string) error {
	return kcat.Container.CopyToContainer(ctx, []byte(data), kcat.FilePath, 700)
}

func (kcat *KcatContainer) ProduceMessageFromFile(ctx context.Context, broker, topic string) error {
	cmd := []string{"kcat", "-b", broker, "-t", topic, "-P", "-l", kcat.FilePath}
	_, _, err := kcat.Container.Exec(ctx, cmd)

	return err
}

func (kcat *KcatContainer) ConsumeMessage(ctx context.Context, broker, topic string) (string, error) {
	cmd := []string{"kcat", "-b", broker, "-C", "-t", topic, "-c1"}
	_, stdout, err := kcat.Container.Exec(ctx, cmd)
	
	if err != nil {
		return "", err
	}

	out, err := io.ReadAll(stdout)

	return string(out), err
}
