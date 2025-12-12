package kafka

import "strings"

const (
	apacheKafkaImagePrefix  = "apache/kafka"
	confluentincImagePrefix = "confluentinc/"
	dockerIoPrefix          = "docker.io/"
)

func isApache(image string) bool {
	return strings.HasPrefix(image, apacheKafkaImagePrefix) || strings.HasPrefix(image, dockerIoPrefix+apacheKafkaImagePrefix)
}

func isConfluentinc(image string) bool {
	return strings.HasPrefix(image, confluentincImagePrefix) || strings.HasPrefix(image, dockerIoPrefix+confluentincImagePrefix)
}
