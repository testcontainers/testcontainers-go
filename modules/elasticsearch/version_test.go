package elasticsearch

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestIsOSS(t *testing.T) {
	assert.Check(t, isOSS("docker.elastic.co/elasticsearch/elasticsearch-oss:latest"))
	assert.Check(t, !isOSS("docker.elastic.co/elasticsearch/elasticsearch:latest"))
}

func TestIsVersion8(t *testing.T) {
	assert.Check(t, isAtLeastVersion("docker.elastic.co/elasticsearch/elasticsearch:latest", 8))

	assert.Check(t, isAtLeastVersion("docker.elastic.co/elasticsearch/elasticsearch:8", 8))
	assert.Check(t, isAtLeastVersion("docker.elastic.co/elasticsearch/elasticsearch:8.0", 8))
	assert.Check(t, isAtLeastVersion("docker.elastic.co/elasticsearch/elasticsearch:8.0.0", 8))
	assert.Check(t, isAtLeastVersion("docker.elastic.co/elasticsearch/elasticsearch:8.1.0", 8))
	assert.Check(t, isAtLeastVersion("docker.elastic.co/elasticsearch/elasticsearch:8.0.1", 8))
	assert.Check(t, isAtLeastVersion("docker.elastic.co/elasticsearch/elasticsearch:9.0.0", 8))
	assert.Check(t, isAtLeastVersion("docker.elastic.co/elasticsearch/elasticsearch:9.0", 8))
	assert.Check(t, isAtLeastVersion("docker.elastic.co/elasticsearch/elasticsearch:9", 8))

	assert.Check(t, !isAtLeastVersion("docker.elastic.co/elasticsearch/elasticsearch:7", 8))
	assert.Check(t, !isAtLeastVersion("docker.elastic.co/elasticsearch/elasticsearch:7.99", 8))
	assert.Check(t, !isAtLeastVersion("docker.elastic.co/elasticsearch/elasticsearch:7.12.99", 8))
	assert.Check(t, !isAtLeastVersion("docker.elastic.co/elasticsearch/elasticsearch:6", 8))
	assert.Check(t, !isAtLeastVersion("docker.elastic.co/elasticsearch/elasticsearch:6.99", 8))
	assert.Check(t, !isAtLeastVersion("docker.elastic.co/elasticsearch/elasticsearch:6.12.99", 8))
}
