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

func getStarterScriptContent(image string) string {
	if isApache(image) {
		return apacheStarterScriptContent
	} else if isConfluentinc(image) {
		return confluentincStarterScriptContent
	} else {
		// Default to confluentinc for backward compatibility
		// in situations when image was custom specified based on confluentinc
		return confluentincStarterScriptContent
	}
}
