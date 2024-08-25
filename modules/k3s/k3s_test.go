package k3s_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kwait "k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/k3s"
	"github.com/testcontainers/testcontainers-go/wait"
)

func Test_LoadImages(t *testing.T) {
	// Give up to three minutes to run this test
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(3*time.Minute))
	defer cancel()

	k3sContainer, err := k3s.Run(ctx, "docker.io/rancher/k3s:v1.27.1-k3s1")
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
	err = provider.PullImage(ctx, "nginx")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Test load image not available", func(t *testing.T) {
		err := k3sContainer.LoadImages(ctx, "fake.registry/fake:non-existing")
		if err == nil {
			t.Fatal("should had failed")
		}
	})

	t.Run("Test load image in cluster", func(t *testing.T) {
		err := k3sContainer.LoadImages(ctx, "nginx")
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

		_, err = k8s.CoreV1().Pods("default").Create(ctx, pod, metav1.CreateOptions{})
		if err != nil {
			t.Fatal(err)
		}

		err = kwait.PollUntilContextCancel(ctx, time.Second, true, func(ctx context.Context) (bool, error) {
			state, err := getTestPodState(ctx, k8s)
			if err != nil {
				return false, err
			}
			if state.Terminated != nil {
				return false, fmt.Errorf("pod terminated: %v", state.Terminated)
			}
			return state.Running != nil, nil
		})
		if err != nil {
			t.Fatal(err)
		}

		state, err := getTestPodState(ctx, k8s)
		if err != nil {
			t.Fatal(err)
		}
		if state.Running == nil {
			t.Fatalf("Unexpected status %v", state)
		}
	})
}

func getTestPodState(ctx context.Context, k8s *kubernetes.Clientset) (corev1.ContainerState, error) {
	var pod *corev1.Pod
	var err error
	pod, err = k8s.CoreV1().Pods("default").Get(ctx, "test-pod", metav1.GetOptions{})
	if err != nil || len(pod.Status.ContainerStatuses) == 0 {
		return corev1.ContainerState{}, err
	}
	return pod.Status.ContainerStatuses[0].State, nil
}

func Test_APIServerReady(t *testing.T) {
	ctx := context.Background()

	k3sContainer, err := k3s.Run(ctx, "docker.io/rancher/k3s:v1.27.1-k3s1")
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

func Test_WithManifestOption(t *testing.T) {
	ctx := context.Background()

	k3sContainer, err := k3s.Run(ctx,
		"docker.io/rancher/k3s:v1.27.1-k3s1",
		k3s.WithManifest("nginx-manifest.yaml"),
		testcontainers.WithWaitStrategy(wait.ForExec([]string{"kubectl", "wait", "pod", "nginx", "--for=condition=Ready"})),
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
}
