package solace

import (
	"reflect"
	"testing"
)

func TestDefaultOptions(t *testing.T) {
	opts := defaultOptions()

	// Test default values
	if opts.vpn != "default" {
		t.Errorf("Expected default vpn to be 'default', got %s", opts.vpn)
	}

	if opts.username != "root" {
		t.Errorf("Expected default username to be 'root', got %s", opts.username)
	}

	if opts.password != "password" {
		t.Errorf("Expected default password to be 'password', got %s", opts.password)
	}

	if opts.shmSize != 1<<30 {
		t.Errorf("Expected default shmSize to be %d, got %d", 1<<30, opts.shmSize)
	}

	// Test that all default service ports are exposed
	expectedPorts := []string{"5672/tcp", "8080/tcp", "55555/tcp", "9000/tcp", "1883/tcp"}
	if len(opts.exposedPorts) != len(expectedPorts) {
		t.Errorf("Expected %d exposed ports, got %d", len(expectedPorts), len(opts.exposedPorts))
	}

	for _, expectedPort := range expectedPorts {
		found := false
		for _, port := range opts.exposedPorts {
			if port == expectedPort {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected port %s to be in exposed ports, but it wasn't found", expectedPort)
		}
	}
}
func TestWithExposedPorts(t *testing.T) {
	tests := []struct {
		name          string
		initialPorts  []string
		newPorts      []string
		expectedPorts []string
	}{
		{
			name:          "add single port",
			initialPorts:  []string{"8080/tcp"},
			newPorts:      []string{"9090/tcp"},
			expectedPorts: []string{"8080/tcp", "9090/tcp"},
		},
		{
			name:          "add multiple ports",
			initialPorts:  []string{"8080/tcp"},
			newPorts:      []string{"9090/tcp", "3000/tcp"},
			expectedPorts: []string{"8080/tcp", "9090/tcp", "3000/tcp"},
		},
		{
			name:          "add to empty ports",
			initialPorts:  []string{},
			newPorts:      []string{"8080/tcp"},
			expectedPorts: []string{"8080/tcp"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &options{exposedPorts: tt.initialPorts}
			option := WithExposedPorts(tt.newPorts...)

			err := option(opts)
			if err != nil {
				t.Errorf("WithExposedPorts returned error: %v", err)
			}

			if !reflect.DeepEqual(opts.exposedPorts, tt.expectedPorts) {
				t.Errorf("Expected exposed ports %v, got %v", tt.expectedPorts, opts.exposedPorts)
			}
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
			if err != nil {
				t.Errorf("WithCredentials returned error: %v", err)
			}

			if opts.username != tt.username {
				t.Errorf("Expected username %s, got %s", tt.username, opts.username)
			}

			if opts.password != tt.password {
				t.Errorf("Expected password %s, got %s", tt.password, opts.password)
			}
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
			option := WithVpn(tt.vpn)

			err := option(opts)
			if err != nil {
				t.Errorf("WithVpn returned error: %v", err)
			}

			if opts.vpn != tt.vpn {
				t.Errorf("Expected vpn %s, got %s", tt.vpn, opts.vpn)
			}
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
			if err != nil {
				t.Errorf("WithQueue returned error: %v", err)
			}

			if opts.queues == nil {
				t.Errorf("Expected queues to be initialized")
				return
			}

			topics, exists := opts.queues[tt.queueName]
			if !exists {
				t.Errorf("Expected queue %s to exist", tt.queueName)
				return
			}

			if len(topics) != tt.expectedLen {
				t.Errorf("Expected %d topics for queue %s, got %d", tt.expectedLen, tt.queueName, len(topics))
			}

			// Check if the new topic is in the list
			found := false
			for _, topic := range topics {
				if topic == tt.topic {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected topic %s to be in queue %s", tt.topic, tt.queueName)
			}
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
			if err != nil {
				t.Errorf("WithShmSize returned error: %v", err)
			}

			if opts.shmSize != tt.expected {
				t.Errorf("Expected shmSize %d, got %d", tt.expected, opts.shmSize)
			}
		})
	}
}

func TestOptionChaining(t *testing.T) {
	// Test that multiple options can be applied together
	opts := defaultOptions()

	// Apply multiple options
	options := []Option{
		WithCredentials("newuser", "newpass"),
		WithVpn("testvpn"),
		WithExposedPorts("3000/tcp", "4000/tcp"),
		WithQueue("queue1", "topic1"),
		WithQueue("queue1", "topic2"),
		WithQueue("queue2", "topic3"),
		WithShmSize(2 << 30),
	}

	for _, option := range options {
		err := option(&opts)
		if err != nil {
			t.Errorf("Option returned error: %v", err)
		}
	}

	// Verify all options were applied
	if opts.username != "newuser" {
		t.Errorf("Expected username 'newuser', got %s", opts.username)
	}
	if opts.password != "newpass" {
		t.Errorf("Expected password 'newpass', got %s", opts.password)
	}
	if opts.vpn != "testvpn" {
		t.Errorf("Expected vpn 'testvpn', got %s", opts.vpn)
	}
	if opts.shmSize != 2<<30 {
		t.Errorf("Expected shmSize %d, got %d", 2<<30, opts.shmSize)
	}

	// Check that custom ports were added
	expectedCustomPorts := []string{"3000/tcp", "4000/tcp"}
	for _, expectedPort := range expectedCustomPorts {
		found := false
		for _, port := range opts.exposedPorts {
			if port == expectedPort {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected custom port %s to be in exposed ports", expectedPort)
		}
	}

	// Check queues
	if len(opts.queues) != 2 {
		t.Errorf("Expected 2 queues, got %d", len(opts.queues))
	}
	if len(opts.queues["queue1"]) != 2 {
		t.Errorf("Expected 2 topics for queue1, got %d", len(opts.queues["queue1"]))
	}
	if len(opts.queues["queue2"]) != 1 {
		t.Errorf("Expected 1 topic for queue2, got %d", len(opts.queues["queue2"]))
	}
}
