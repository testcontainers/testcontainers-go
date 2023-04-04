# Kafka

Testcontainers can automatically start an Apache Kafka container in KRaft (Kafka Raft)
mode. The container uses the official Confluent Kafka Platform image.

## Adding this module to your project dependencies

Please run the following command to add the Kafka module to your Go dependencies:

```
go get github.com/testcontainers/testcontainers-go/modules/kafka
```

## Usage example

Create a `Kafka` container to use it in your tests:

<!--codeinclude-->
[Creating a Kafka container](../../modules/kafka/kafka.go)
<!--/codeinclude-->

After startup of the container, the broker address can be retrieved as following:

```go
brokers, err := container.Brokers(context.Background())
if err != nil {
    // ... error handling
}
```

A full example with Sarama producer and consumer:

<!--codeinclude-->
[Test for a Kafka container](../../modules/kafka/kafka_test.go) inside_block:TestStartContainer
<!--/codeinclude-->
