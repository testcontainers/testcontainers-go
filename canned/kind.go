package canned

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"time"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/go-connections/nat"
	"github.com/pkg/errors"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cluster/config/encoding"
	"sigs.k8s.io/kind/pkg/cluster/create"
	"sigs.k8s.io/kind/pkg/util"
)

// seededRand provides the rand function to calculate a different name for
// every kind cluster.
var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

// namePrefix is the fixed part that the container name created by kind will
// have
const namePrefix = "testcontainers-"

// dataPlanFromKind is the suffix kind attaches to all the containers
const dataPlanFromKind = "-control-plane"

// Implement interfaces
var _ testcontainers.Container = (*KubeKindContainer)(nil)

// KubeKindContainer is the struct that spins up a Kubernetes Cluster via
// Kind.
type KubeKindContainer struct {
	container   testcontainers.Container
	kindContext *cluster.Context
	clientset   *kubernetes.Clientset
	clusterName string
}

func (k *KubeKindContainer) Endpoint(ctx context.Context, endpoint string) (string, error) {
	return k.container.Endpoint(ctx, endpoint)
}

func (k *KubeKindContainer) PortEndpoint(ctx context.Context, port nat.Port, s string) (string, error) {
	return k.container.PortEndpoint(ctx, port, s)
}

func (k *KubeKindContainer) Host(ctx context.Context) (string, error) {
	return k.container.Host(ctx)
}

func (k *KubeKindContainer) MappedPort(ctx context.Context, p nat.Port) (nat.Port, error) {
	return k.container.MappedPort(ctx, p)
}

func (k *KubeKindContainer) Ports(ctx context.Context) (nat.PortMap, error) {
	return k.container.Ports(ctx)
}

func (k *KubeKindContainer) SessionID() string {
	return k.container.SessionID()
}

// GetContainerID returns the container ID from Docker
func (k *KubeKindContainer) GetContainerID() string {
	return k.container.GetContainerID()
}

// GetClientset returns a configured Kubernetes Clientset ready to be used.
func (k *KubeKindContainer) GetClientset() (*kubernetes.Clientset, error) {
	if k.clientset != nil {
		return k.clientset, nil
	}
	config, err := clientcmd.BuildConfigFromFlags("", k.kindContext.KubeConfigPath())
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	k.clientset = clientset
	return k.clientset, nil
}

// Start creates and runs the container.
func (k *KubeKindContainer) Start(ctxMain context.Context) error {
	renameAttempt := 0
rename:
	k.clusterName = fmt.Sprintf("%s%d", namePrefix, seededRand.Int())
	kubeconfig := ""
	cfg, err := encoding.Load(kubeconfig)
	if err != nil {
		return errors.Wrap(err, "error loading config")
	}

	// validate the config.
	err = cfg.Validate()
	if err != nil {
		configErrors := err.(util.Errors)
		return configErrors.Errors()[0]
	}

	known, err := cluster.IsKnown(k.clusterName)
	if err != nil {
		return err
	}
	if known {
		if renameAttempt > 9 {
			return errors.Errorf("a cluster with the name %q already exists", k.clusterName)
		}
		renameAttempt = renameAttempt + 1
		goto rename
	}
	k.kindContext = cluster.NewContext(k.clusterName)
	if err = k.kindContext.Create(cfg,
		create.Retain(false),
		create.WaitForReady(time.Minute*5),
	); err != nil {
		return errors.Wrap(err, "failed to create cluster")
	}
	args := filters.NewArgs(filters.Arg("name", k.clusterName+dataPlanFromKind))
	if err != nil {
		return err
	}
	p, err := testcontainers.NewDockerProvider()
	if err != nil {
		return err
	}
	c, err := p.ContainerFromDockerArgs(ctxMain, args)
	if err != nil {
		return err
	}
	k.container = c
	return nil
}

// Logs returns a Reader with the logs from the container.
func (k *KubeKindContainer) Logs(ctx context.Context) (io.ReadCloser, error) {
	return k.container.Logs(ctx)
}

// Name returns the name of Kind cluster (not the container).
func (k *KubeKindContainer) Name(ctxMain context.Context) (string, error) {
	return k.clusterName, nil
}

// Terminate deletes the cluster via Kind.
func (k *KubeKindContainer) Terminate(ctxMain context.Context) error {
	if err := k.kindContext.Delete(); err != nil {
		return errors.Wrap(err, "failed to delete cluster")
	}
	return nil
}
