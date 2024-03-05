package core_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/internal/core"
)

func TestGetDockerHostIPs(t *testing.T) {
	type args struct {
		host string
	}
	tests := []struct {
		name    string
		args    args
		hostIps []core.HostIP
	}{
		{
			name: "should return a list of resolved host IPs when host is not an IP",
			args: args{
				host: "localhost",
			},
			hostIps: []core.HostIP{{Address: "127.0.0.1", Family: core.IPv4}},
		},
		{
			name: "should return host IP and v4 family when host is an IPv4 IP",
			args: args{
				host: "127.0.0.1",
			},
			hostIps: []core.HostIP{{Address: "127.0.0.1", Family: core.IPv4}},
		},
		{
			name: "should return host IP and v4 family when host is an IPv4 IP with tcp schema",
			args: args{
				host: "tcp://127.0.0.1:64692",
			},
			hostIps: []core.HostIP{{Address: "127.0.0.1", Family: core.IPv4}},
		},
		{
			name: "should return host IP and v6 family when host is an IPv6 IP",
			args: args{
				host: "::1",
			},
			hostIps: []core.HostIP{{Address: "::1", Family: core.IPv6}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hips := core.GetDockerHostIPs(tt.args.host)
			assert.Equal(t, tt.hostIps, hips)
		})
	}
}
