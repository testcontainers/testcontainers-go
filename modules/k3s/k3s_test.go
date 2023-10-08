package k3s_test

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/k3s"
)

func Test_LoadImages(t *testing.T) {
	ctx := context.Background()

	k3sContainer, err := k3s.RunContainer(ctx,
		testcontainers.WithImage("docker.io/rancher/k3s:v1.27.1-k3s1"),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container
	defer func() {
		if err := k3sContainer.Terminate(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	kubeConfigYaml, err := k3sContainer.GetKubeConfig(ctx)
	if err != nil {
		t.Fatal(err)
	}

	restcfg, err := clientcmd.RESTConfigFromKubeConfig(kubeConfigYaml)
	if err != nil {
		t.Fatal(err)
	}

	k8s, err := kubernetes.NewForConfig(restcfg)
	if err != nil {
		t.Fatal(err)
	}

	provider, err := testcontainers.ProviderDocker.GetProvider()
	if err != nil {
		t.Fatal(err)
	}

	// ensure nginx image is available locally
	err = provider.PullImage(context.Background(), "nginx")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Test load image not available", func(t *testing.T) {
		err := k3sContainer.LoadImages(context.Background(), "fake.registry/fake:non-existing")
		if err == nil {
			t.Fatal("should had failed")
		}
	})

	t.Run("Test load image in cluster", func(t *testing.T) {
		err := k3sContainer.LoadImages(context.Background(), "nginx")
		if err != nil {
			t.Fatal(err)
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
			t.Fatal(err)
		}

		time.Sleep(1 * time.Second)
		pod, err = k8s.CoreV1().Pods("default").Get(context.Background(), "test-pod", metav1.GetOptions{})
		if err != nil {
			t.Fatal(err)
		}
		waiting := pod.Status.ContainerStatuses[0].State.Waiting
		if waiting != nil && waiting.Reason == "ErrImageNeverPull" {
			t.Fatal("Image was not loaded")
		}
	})
}

func Test_APIServerReady(t *testing.T) {
	ctx := context.Background()

	k3sContainer, err := k3s.RunContainer(ctx,
		testcontainers.WithImage("docker.io/rancher/k3s:v1.27.1-k3s1"),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container
	defer func() {
		if err := k3sContainer.Terminate(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	kubeConfigYaml, err := k3sContainer.GetKubeConfig(ctx)
	if err != nil {
		t.Fatal(err)
	}

	restcfg, err := clientcmd.RESTConfigFromKubeConfig(kubeConfigYaml)
	if err != nil {
		t.Fatal(err)
	}

	k8s, err := kubernetes.NewForConfig(restcfg)
	if err != nil {
		t.Fatal(err)
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
					Name:  "nginx",
					Image: "nginx",
				},
			},
		},
	}

	_, err = k8s.CoreV1().Pods("default").Create(context.Background(), pod, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create pod %v", err)
	}
}
