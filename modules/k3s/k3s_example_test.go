package k3s_test

import (
	"context"
	"fmt"
	"log"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/testcontainers/testcontainers-go/modules/k3s"
)

func ExampleRun() {
	// runK3sContainer {
	ctx := context.Background()

	k3sContainer, err := k3s.Run(ctx, "docker.io/rancher/k3s:v1.27.1-k3s1")
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := k3sContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	// }

	state, err := k3sContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	kubeConfigYaml, err := k3sContainer.GetKubeConfig(ctx)
	if err != nil {
		log.Fatalf("failed to get kubeconfig: %s", err)
	}

	restcfg, err := clientcmd.RESTConfigFromKubeConfig(kubeConfigYaml)
	if err != nil {
		log.Fatalf("failed to create rest config: %s", err)
	}

	k8s, err := kubernetes.NewForConfig(restcfg)
	if err != nil {
		log.Fatalf("failed to create k8s client: %s", err)
	}

	nodes, err := k8s.CoreV1().Nodes().List(ctx, v1.ListOptions{})
	if err != nil {
		log.Fatalf("failed to list nodes: %s", err)
	}

	fmt.Println(len(nodes.Items))

	// Output:
	// true
	// 1
}
