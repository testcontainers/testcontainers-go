package k3s

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func TestK3s(t *testing.T) {
	ctx := context.Background()

	container, err := RunContainer(ctx,
		testcontainers.WithWaitStrategy(wait.ForLog("Starting node config controller")))
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
	kubeConfigYaml, err := container.getkubeConfigYaml(ctx)
	if err != nil {
		t.Fatalf("failed to get kube-config : %s", err)
	}
	fmt.Println("---from test---")
	fmt.Println(kubeConfigYaml)

	restcfg, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeConfigYaml))
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
	fmt.Println(nodes)

	assert.Equal(t, len(nodes.Items), 1)
}

// func TestK3sKubectlContainer(t *testing.T) {
// 	ctx := context.Background()

// 	nw, err := testcontainers.GenericNetwork(ctx, testcontainers.GenericNetworkRequest{
// 		NetworkRequest: testcontainers.NetworkRequest{
// 			Name: "k3s-network",
// 		},
// 	})

// 	require.Nil(t, err)
// 	assert.NotNil(t, nw)
// 	networks := []string{"k3s-network"}
// 	networkAlias := map[string][]string{
// 		networks[0]: {"k3s"},
// 	}
// 	container, err := RunContainer(ctx,
// 		WithNetwork(networks),
// 		WithNetworkAlias(networkAlias),
// 	)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	// Clean up the container after the test is complete
// 	t.Cleanup(func() {
// 		if err := container.Terminate(ctx); err != nil {
// 			t.Fatalf("failed to terminate container: %s", err)
// 		}
// 	})

// 	// perform assertions
// 	kubeConfigYaml, err := container.getInternalKubeConfigYaml(ctx, "k3s-network")
// 	if err != nil {
// 		t.Fatalf("failed to get kube-config : %s", err)
// 	}
// 	fmt.Println("---from test---")
// 	fmt.Println(kubeConfigYaml)

// 	if err := os.WriteFile(filepath.Join(".", "kubeconfig.yaml"), []byte(kubeConfigYaml), 0755); err != nil {
// 		t.Errorf("Failed to create the file: %v", err)
// 		return
// 	}

// 	kubectlContainer, err := RunContainer(ctx, testcontainers.WithImage("rancher/kubectl:v1.27.1"),
// 		WithNetwork(networks),
// 		WithKubectlConfigFile(filepath.Join(".", "kubeconfig.yaml")),
// 		WithCmd([]string{"get namespaces"}),
// 	)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	// Clean up the container after the test is complete
// 	t.Cleanup(func() {
// 		if err := kubectlContainer.Terminate(ctx); err != nil {
// 			t.Fatalf("failed to terminate container: %s", err)
// 		}
// 	})
// 	r, err := kubectlContainer.Logs(ctx)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer r.Close()
// 	b, err := io.ReadAll(r)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	assert.Contains(t, string(b), "kube-system")
// }
