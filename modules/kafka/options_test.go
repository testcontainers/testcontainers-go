package kafka

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
			want:  ApacheStarterScript,
		},
		{
			name:  "apache native image - specific version",
			image: "apache/kafka-native:4.0.1",
			want:  ApacheStarterScript,
		},
		{
			name:  "apache native image - specific version with docker.io prefix",
			image: "docker.io/apache/kafka-native:4.0.1",
			want:  ApacheStarterScript,
		},
		{
			name:  "confluentinc image - latest",
			image: "confluentinc/cp-kafka:latest",
			want:  ConfluentStarterScript,
		},
		{
			name:  "confluentinc image - no tag",
			image: "confluentinc/cp-kafka",
			want:  ConfluentStarterScript,
		},
		{
			name:  "confluentinc image - specific version",
			image: "confluentinc/cp-kafka:8.1.0",
			want:  ConfluentStarterScript,
		},
		{
			name:  "confluentinc image - specific version with docker.io prefix",
			image: "docker.io/confluentinc/cp-kafka:8.1.0",
			want:  ConfluentStarterScript,
		},
		{
			name:  "custom image",
			image: "custom/kafka:latest",
			want:  ConfluentStarterScript,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &runOptions{
				image: tt.image,
			}
			if got := opts.getStarterScriptContent(); got != tt.want {
				t.Errorf("getStarterScriptContent() = %v, want %v", got, tt.want)
			}

			assert.NoError(t, WithStarterScript("mytestsript")(opts))
			if got := opts.getStarterScriptContent(); got != "mytestsript" {
				t.Errorf("getStarterScriptContent() with explicit setting = %v, want %v", got, "mytestsript")
			}
		})
	}
}
