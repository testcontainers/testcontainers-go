package k3s_test

import (
	"context"
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/k3s"
)

func ExampleRunContainer() {
	// runK3sContainer {
	ctx := context.Background()

	k3sContainer, err := k3s.RunContainer(ctx,
		testcontainers.WithImage("docker.io/rancher/k3s:v1.27.1-k3s1"),
	)
	if err != nil {
		panic(err)
	}

	// Clean up the container
	defer func() {
		if err := k3sContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()
	// }

	state, err := k3sContainer.State(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println(state.Running)

	kubeConfigYaml, err := k3sContainer.GetKubeConfig(ctx)
	if err != nil {
		panic(err)
	}

	restcfg, err := clientcmd.RESTConfigFromKubeConfig(kubeConfigYaml)
	if err != nil {
		panic(err)
	}

	k8s, err := kubernetes.NewForConfig(restcfg)
	if err != nil {
		panic(err)
	}

	nodes, err := k8s.CoreV1().Nodes().List(ctx, v1.ListOptions{})
	if err != nil {
		panic(err)
	}

	fmt.Println(len(nodes.Items))

	// Output:
	// true
	// 1
}
