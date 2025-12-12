package kafka

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_isApache(t *testing.T) {
	tests := []struct {
		name  string
		image string
		want  bool
	}{
		{
			name:  "apache native image - no tag",
			image: "apache/kafka-native",
			want:  true,
		},
		{
			name:  "apache native image - latest",
			image: "apache/kafka-native:latest",
			want:  true,
		},
		{
			name:  "apache native image - specific version",
			image: "apache/kafka-native:4.0.1",
			want:  true,
		},
		{
			name:  "apache native image - specific version with docker.io prefix",
			image: "docker.io/apache/kafka-native:4.0.1",
			want:  true,
		},
		{
			name:  "apache not-native image - no tag",
			image: "apache/kafka",
			want:  true,
		},
		{
			name:  "apache not-native image - latest",
			image: "apache/kafka:latest",
			want:  true,
		},
		{
			name:  "apache not-native image - specific version",
			image: "apache/kafka:4.0.1",
			want:  true,
		},
		{
			name:  "apache not-native image - specific version with docker.io prefix",
			image: "docker.io/apache/kafka:4.0.1",
			want:  true,
		},
		{
			name:  "confluentinc image",
			image: "confluentinc/cp-kafka:latest",
			want:  false,
		},
		{
			name:  "custom image",
			image: "custom/kafka:latest",
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, isApache(tt.image))
		})
	}
}

func Test_isConfluentinc(t *testing.T) {
	tests := []struct {
		name  string
		image string
		want  bool
	}{
		{
			name:  "confluentinc image - no tag",
			image: "confluentinc/cp-kafka",
			want:  true,
		},
		{
			name:  "confluentinc image - latest",
			image: "confluentinc/cp-kafka:latest",
			want:  true,
		},
		{
			name:  "confluentinc image - specific version",
			image: "confluentinc/cp-kafka:8.1.0",
			want:  true,
		},
		{
			name:  "confluentinc image - specific version with docker.io prefix",
			image: "docker.io/confluentinc/cp-kafka:8.1.0",
			want:  true,
		},
		{
			name:  "apache native image",
			image: "apache/kafka-native:latest",
			want:  false,
		},
		{
			name:  "custom image",
			image: "custom/kafka:latest",
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, isConfluentinc(tt.image))
		})
	}
}
