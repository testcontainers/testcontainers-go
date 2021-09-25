// inspired by Java Kafka testcontainers' module
//https://github.com/testcontainers/testcontainers-java/blob/master/modules/kafka/src/main/java/org/testcontainers/containers/KafkaContainer.java

package canned

import (
	"context"
	"github.com/testcontainers/testcontainers-go"
	"io/ioutil"
	"os"
)

const (
	clusterName     = "kafka-cluster"
	zookeeperPort   = "2181"
	kafkaBrokerPort = "9092"
	kafkaClientPort = "9093"
	zookeeperImage  = "confluentinc/cp-zookeeper:5.2.1"
	kafkaImage      = "confluentinc/cp-kafka:5.2.1"
)

type KafkaCluster struct {
	kafkaContainer     testcontainers.Container
	zookeeperContainer testcontainers.Container
}

// StartCluster starts kafka cluster
func (kc *KafkaCluster) StartCluster() {
	ctx := context.Background()

	kc.zookeeperContainer.Start(ctx)
	kc.kafkaContainer.Start(ctx)
	kc.startKafka()
}

// GetKafkaHost gets the kafka host:port so it can be accessed from outside the container
func (kc *KafkaCluster) GetKafkaHost() string {
	ctx := context.Background()
	host, err := kc.kafkaContainer.Host(ctx)
	if err != nil {
		panic(err)
	}
	port, err := kc.kafkaContainer.MappedPort(ctx, kafkaClientPort)
	if err != nil {
		panic(err)
	}

	// returns the exposed kafka host:port
	return host + ":" + port.Port()
}

func (kc *KafkaCluster) startKafka() {
	ctx := context.Background()

	kafkaStartFile, err := ioutil.TempFile("", "testcontainers_start.sh")
	if err != nil {
		panic(err)
	}
	defer os.Remove(kafkaStartFile.Name())

	// needs to set KAFKA_ADVERTISED_LISTENERS with the exposed kafka port
	exposedHost := kc.GetKafkaHost()
	kafkaStartFile.WriteString("#!/bin/bash \n")
	kafkaStartFile.WriteString("export KAFKA_ADVERTISED_LISTENERS='PLAINTEXT://" + exposedHost + ",BROKER://kafka:" + kafkaBrokerPort + "'\n")
	kafkaStartFile.WriteString(". /etc/confluent/docker/bash-config \n")
	kafkaStartFile.WriteString("/etc/confluent/docker/configure \n")
	kafkaStartFile.WriteString("/etc/confluent/docker/launch \n")

	err = kc.kafkaContainer.CopyFileToContainer(ctx, kafkaStartFile.Name(), "testcontainers_start.sh", 0700)
	if err != nil {
		panic(err)
	}
}

func NewKafkaCluster() *KafkaCluster {
	ctx := context.Background()

	// creates a network, so kafka and zookeeper can communicate directly
	network, err := testcontainers.GenericNetwork(ctx, testcontainers.GenericNetworkRequest{
		NetworkRequest: testcontainers.NetworkRequest{Name: clusterName},
	})
	if err != nil {
		panic(err)
	}

	dockerNetwork := network.(*testcontainers.DockerNetwork)

	zookeeperContainer := createZookeeperContainer(dockerNetwork)
	kafkaContainer := createKafkaContainer(dockerNetwork)

	return &KafkaCluster{
		zookeeperContainer: zookeeperContainer,
		kafkaContainer:     kafkaContainer,
	}
}

func createZookeeperContainer(network *testcontainers.DockerNetwork) testcontainers.Container {
	ctx := context.Background()

	// creates the zookeeper container, but do not start it yet
	zookeeperContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:          zookeeperImage,
			ExposedPorts:   []string{zookeeperPort},
			Env:            map[string]string{"ZOOKEEPER_CLIENT_PORT": zookeeperPort, "ZOOKEEPER_TICK_TIME": "2000"},
			Networks:       []string{network.Name},
			NetworkAliases: map[string][]string{network.Name: {"zookeeper"}},
		},
	})
	if err != nil {
		panic(err)
	}

	return zookeeperContainer
}

func createKafkaContainer(network *testcontainers.DockerNetwork) testcontainers.Container {
	ctx := context.Background()

	// creates the kafka container, but do not start it yet
	kafkaContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        kafkaImage,
			ExposedPorts: []string{kafkaClientPort},
			Env: map[string]string{
				"KAFKA_BROKER_ID":                        "1",
				"KAFKA_ZOOKEEPER_CONNECT":                "zookeeper:" + zookeeperPort,
				"KAFKA_LISTENERS":                        "PLAINTEXT://0.0.0.0:" + kafkaClientPort + ",BROKER://0.0.0.0:" + kafkaBrokerPort,
				"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP":   "BROKER:PLAINTEXT,PLAINTEXT:PLAINTEXT",
				"KAFKA_INTER_BROKER_LISTENER_NAME":       "BROKER",
				"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR": "1",
			},
			Networks:       []string{network.Name},
			NetworkAliases: map[string][]string{network.Name: {"kafka"}},
			// the container only starts when it finds and run /testcontainers_start.sh
			Cmd: []string{"sh", "-c", "while [ ! -f /testcontainers_start.sh ]; do sleep 0.1; done; /testcontainers_start.sh"},
		},
	})
	if err != nil {
		panic(err)
	}

	return kafkaContainer
}
