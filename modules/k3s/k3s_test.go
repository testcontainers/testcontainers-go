package k3s

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestK3s(t *testing.T) {
	ctx := context.Background()

	// k3sRunContainer {
	container, err := RunContainer(ctx,
		testcontainers.WithWaitStrategy(wait.ForLog("Starting node config controller")))
	if err != nil {
		t.Fatal(err)
	}
	// }

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// GetKubeConfig {
	kubeConfigYaml, err := container.GetKubeConfig(ctx)
	if err != nil {
		t.Fatalf("failed to get kube-config : %s", err)
	}
	// }

	restcfg, err := clientcmd.RESTConfigFromKubeConfig(kubeConfigYaml)
	if err != nil {
		t.Fatalf("failed to create rest client for kubernetes : %s", err)
	}

	k8s, err := kubernetes.NewForConfig(restcfg)
	if err != nil {
		t.Fatalf("failed to place config in k8s clientset : %s", err)
	}

	nodes, err := k8s.CoreV1().Nodes().List(ctx, v1.ListOptions{})
	if err != nil {
		t.Fatalf("failed to get list of nodes : %s", err)
	}

	assert.Equal(t, len(nodes.Items), 1)
}
