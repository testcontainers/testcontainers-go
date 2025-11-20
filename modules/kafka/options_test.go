package kafka

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_runOptions_getStarterScriptContent(t *testing.T) {
	tests := []struct {
		name  string
		image string
		want  string
	}{
		{
			name:  "apache native image - latest",
			image: "apache/kafka-native:latest",
			want:  apacheStarterScript,
		},
		{
			name:  "apache native image - specific version",
			image: "apache/kafka-native:4.0.1",
			want:  apacheStarterScript,
		},
		{
			name:  "apache native image - specific version with docker.io prefix",
			image: "docker.io/apache/kafka-native:4.0.1",
			want:  apacheStarterScript,
		},
		{
			name:  "confluentinc image - latest",
			image: "confluentinc/cp-kafka:latest",
			want:  confluentStarterScript,
		},
		{
			name:  "confluentinc image - no tag",
			image: "confluentinc/cp-kafka",
			want:  confluentStarterScript,
		},
		{
			name:  "confluentinc image - specific version",
			image: "confluentinc/cp-kafka:8.1.0",
			want:  confluentStarterScript,
		},
		{
			name:  "confluentinc image - specific version with docker.io prefix",
			image: "docker.io/confluentinc/cp-kafka:8.1.0",
			want:  confluentStarterScript,
		},
		{
			name:  "custom image",
			image: "custom/kafka:latest",
			want:  confluentStarterScript,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &runOptions{
				image: tt.image,
			}
			require.Equal(t, tt.want, opts.getStarterScriptContent())
			require.NoError(t, WithStarterScript("mytestsript")(opts))
			require.Equal(t, "mytestsript", opts.getStarterScriptContent())
		})
	}
}
