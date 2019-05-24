package canned

import (
	"context"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestKindCreationAndTermination(t *testing.T) {
	ctx := context.Background()
	k := &KubeKindContainer{}
	err := k.Start(ctx)
	if err != nil {
		t.Fatal(err.Error())
	}
	time.Sleep(2 * time.Second)
	err = k.Terminate(ctx)
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestGetDefaultNamespace(t *testing.T) {
	ctx := context.Background()
	k := &KubeKindContainer{}
	err := k.Start(ctx)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer k.Terminate(ctx)
	clientset, err := k.GetClientset()
	if err != nil {
		t.Fatal(err.Error())
	}
	ns, err := clientset.CoreV1().Namespaces().Get("default", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err.Error())
	}
	if ns.GetName() != "default" {
		t.Fatalf("Expected default namespace got %s", ns.GetName())
	}
}

func TestGetLogs(t *testing.T) {
	ctx := context.Background()
	k := &KubeKindContainer{}
	err := k.Start(ctx)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer k.Terminate(ctx)
	reader, err := k.Logs(ctx)
	if err != nil {
		t.Fatal(err.Error())
	}
	b, err := ioutil.ReadAll(reader)
	logs := string(b)
	if !strings.Contains(logs, "Initializing machine ID from random generator") {
		t.Fatal("Didn't find the sentence \"Initializing machine ID from random generator\"")
	}
}

func ExampleKubeKindContainer_Start() {
	ctx := context.Background()
	k := &KubeKindContainer{}
	err := k.Start(ctx)
	if err != nil {
		panic(err.Error())
	}
}

func ExampleKubeKindContainer_Terminate() {
	ctx := context.Background()
	k := &KubeKindContainer{}
	err := k.Start(ctx)
	if err != nil {
		panic(err.Error())
	}
	defer k.Terminate(ctx)
}

func ExampleKubeKindContainer_Logs() {
	ctx := context.Background()
	k := &KubeKindContainer{}
	err := k.Start(ctx)
	if err != nil {
		panic(err.Error())
	}
	defer k.Terminate(ctx)

	reader, err := k.Logs(ctx)
	if err != nil {
		panic(err.Error())
	}
	b, err := ioutil.ReadAll(reader)
	logs := string(b)
	if !strings.Contains(logs, "Reached target") {
		panic("Reached target")
	}
}
