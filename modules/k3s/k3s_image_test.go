package k3s_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/k3s"
	"github.com/testcontainers/testcontainers-go/wait"

)

func ExampleLoadImages() {
	ctx := context.Background()

	k3sContainer, err := k3s.RunContainer(ctx,
		testcontainers.WithImage("docker.io/rancher/k3s:v1.27.1-k3s1"),
		testcontainers.WithWaitStrategy(wait.ForLog(".*Node controller sync successful.*").AsRegexp()),
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

	provider, err := testcontainers.ProviderDocker.GetProvider()
	if err != nil {
		panic(err)
	}

	err = provider.PullImage(context.Background(), "nginx")
	if err != nil {
		panic(err)
	}

	output := filepath.Join(os.TempDir(), "nginx.tar")
	err = provider.SaveImages(context.Background(), output, "nginx")
	if err != nil {
		panic(err)
	}

	err = k3sContainer.LoadImages(context.Background(), output)
	if err != nil {
		panic(err)
	}

	pod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
		Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:            "nginx",
					Image:           "nginx",
					ImagePullPolicy: corev1.PullNever, // use image only if already present
				},
			},
		},
	}

	_, err = k8s.CoreV1().Pods("default").Create(context.Background(), pod, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}

	time.Sleep(1 * time.Second)
	pod, err = k8s.CoreV1().Pods("default").Get(context.Background(), "test-pod", metav1.GetOptions{})
	if err != nil {
		panic(err)
	}

	if pod.Status.ContainerStatuses[0].State.Waiting.Reason == "ErrImageNeverPull" {
		panic(fmt.Errorf("Image was not loaded"))
	}
}
