package solace

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultOptions(t *testing.T) {
	opts := defaultOptions()

	// Test default values
	assert.Equal(t, "default", opts.vpn)
	assert.Equal(t, "root", opts.username)
	assert.Equal(t, "password", opts.password)
	assert.Equal(t, int64(1<<30), opts.shmSize)

	// Test that all default services are included
	expectedServices := []Service{ServiceAMQP, ServiceSMF, ServiceREST, ServiceMQTT}
	require.Len(t, opts.services, len(expectedServices))

	for _, expectedService := range expectedServices {
		found := false
		for _, service := range opts.services {
			if service.Name == expectedService.Name && service.Port == expectedService.Port {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected service %s:%d to be in services", expectedService.Name, expectedService.Port)
	}
}

func TestWithServices(t *testing.T) {
	tests := []struct {
		name             string
		services         []Service
		expectedServices []Service
	}{
		{
			name:             "set single service",
			services:         []Service{ServiceAMQP},
			expectedServices: []Service{ServiceAMQP},
		},
		{
			name:             "set multiple services",
			services:         []Service{ServiceAMQP, ServiceMQTT, ServiceREST},
			expectedServices: []Service{ServiceAMQP, ServiceMQTT, ServiceREST},
		},
		{
			name:             "set no services",
			services:         []Service{},
			expectedServices: []Service{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &options{}
			option := WithServices(tt.services...)

			err := option(opts)
			if len(tt.services) == 0 {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedServices, opts.services)
		})
	}
}

func TestWithCredentials(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
	}{
		{
			name:     "valid credentials",
			username: "testuser",
			password: "testpass",
		},
		{
			name:     "empty credentials",
			username: "",
			password: "",
		},
		{
			name:     "special characters",
			username: "user@domain.com",
			password: "p@$$w0rd!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &options{}
			option := WithCredentials(tt.username, tt.password)

			err := option(opts)
			require.NoError(t, err)

			assert.Equal(t, tt.username, opts.username)
			assert.Equal(t, tt.password, opts.password)
		})
	}
}

func TestWithVpn(t *testing.T) {
	tests := []struct {
		name string
		vpn  string
	}{
		{
			name: "valid vpn name",
			vpn:  "myvpn",
		},
		{
			name: "empty vpn name",
			vpn:  "",
		},
		{
			name: "vpn with special characters",
			vpn:  "my-vpn_123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &options{}
			option := WithVPN(tt.vpn)

			err := option(opts)
			require.NoError(t, err)

			assert.Equal(t, tt.vpn, opts.vpn)
		})
	}
}

func TestWithQueue(t *testing.T) {
	tests := []struct {
		name        string
		queueName   string
		topic       string
		existing    map[string][]string
		expectedLen int
	}{
		{
			name:        "add topic to new queue",
			queueName:   "testqueue",
			topic:       "testtopic",
			existing:    nil,
			expectedLen: 1,
		},
		{
			name:        "add topic to existing queue",
			queueName:   "testqueue",
			topic:       "newtopic",
			existing:    map[string][]string{"testqueue": {"oldtopic"}},
			expectedLen: 2,
		},
		{
			name:        "add topic to different queue",
			queueName:   "newqueue",
			topic:       "topic1",
			existing:    map[string][]string{"existingqueue": {"topic2"}},
			expectedLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &options{queues: tt.existing}
			option := WithQueue(tt.queueName, tt.topic)

			err := option(opts)
			require.NoError(t, err)

			assert.NotNil(t, opts.queues, "Expected queues to be initialized")

			topics, exists := opts.queues[tt.queueName]
			assert.True(t, exists, "Expected queue %s to exist", tt.queueName)

			assert.Len(t, topics, tt.expectedLen, "Expected %d topics for queue %s", tt.expectedLen, tt.queueName)

			// Check if the new topic is in the list
			assert.Contains(t, topics, tt.topic, "Expected topic %s to be in queue %s", tt.topic, tt.queueName)
		})
	}
}

func TestWithShmSize(t *testing.T) {
	tests := []struct {
		name     string
		shmSize  int64
		expected int64
	}{
		{
			name:     "valid shm size",
			shmSize:  2 << 30, // 2 GiB
			expected: 2 << 30,
		},
		{
			name:     "zero shm size",
			shmSize:  0,
			expected: 0,
		},
		{
			name:     "small shm size",
			shmSize:  1024,
			expected: 1024,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &options{}
			option := WithShmSize(tt.shmSize)

			err := option(opts)
			require.NoError(t, err)

			assert.Equal(t, tt.expected, opts.shmSize)
		})
	}
}

func TestOptionChaining(t *testing.T) {
	// Test that multiple options can be applied together
	opts := defaultOptions()

	// Apply multiple options
	options := []Option{
		WithCredentials("newuser", "newpass"),
		WithVPN("testvpn"),
		WithServices(ServiceAMQP, ServiceMQTT), // Use WithServices instead of WithExposedPorts
		WithQueue("queue1", "topic1"),
		WithQueue("queue1", "topic2"),
		WithQueue("queue2", "topic3"),
		WithShmSize(2 << 30),
	}

	for _, option := range options {
		err := option(&opts)
		require.NoError(t, err)
	}

	// Verify all options were applied
	assert.Equal(t, "newuser", opts.username)
	assert.Equal(t, "newpass", opts.password)
	assert.Equal(t, "testvpn", opts.vpn)
	assert.Equal(t, int64(2<<30), opts.shmSize)

	// Check that services were set correctly
	expectedServices := []Service{ServiceAMQP, ServiceMQTT}
	require.Len(t, opts.services, len(expectedServices))
	for i, expectedService := range expectedServices {
		assert.Equal(t, expectedService.Name, opts.services[i].Name)
		assert.Equal(t, expectedService.Port, opts.services[i].Port)
	}

	// Check queues
	assert.Len(t, opts.queues, 2)
	assert.Len(t, opts.queues["queue1"], 2)
	assert.Len(t, opts.queues["queue2"], 1)
}
