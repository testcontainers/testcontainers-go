package sbx

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	portLeaseLabel     = "org.testcontainers.sbx.port-lease"
	portPublishTries   = 10
	portCommandTimeout = 30 * time.Second
)

type portBinding struct {
	HostIP   string `json:"HostIp"`
	HostPort string `json:"HostPort"`
}

type portBindings map[string][]portBinding

type preparedCreate struct {
	Token    string
	Bindings portBindings
}

type portPublication struct {
	HostIP      string `json:"host_ip"`
	HostPort    int    `json:"host_port"`
	SandboxPort int    `json:"sandbox_port"`
	Protocol    string `json:"protocol"`
}

func (p portPublication) spec() string {
	return net.JoinHostPort(p.HostIP, strconv.Itoa(p.HostPort)) + ":" + strconv.Itoa(p.SandboxPort) + "/" + p.Protocol
}

type portPlan struct {
	containerPort string
	publication   portPublication
	ephemeral     bool
}

func planBindings(ctx context.Context, bindings portBindings, forbidden map[int]bool) ([]portPlan, error) {
	keys := make([]string, 0, len(bindings))
	for key := range bindings {
		keys = append(keys, key)
	}
	slices.Sort(keys)

	plans := make([]portPlan, 0)
	selected := make(map[int]bool)
	for _, key := range keys {
		_, protocol, ok := strings.Cut(key, "/")
		containerPort, parseErr := parseContainerPort(key)
		if !ok || parseErr != nil || (protocol != "tcp" && protocol != "udp") {
			return nil, fmt.Errorf("invalid Docker container port %q: sbx supports one TCP or UDP port", key)
		}
		_ = containerPort // The Docker key is retained; sbx targets the chosen equal host/guest port.
		for _, binding := range bindings[key] {
			if err := validateHostIP(binding.HostIP); err != nil {
				return nil, fmt.Errorf("container port %s: %w", key, err)
			}
			hostPort, ephemeral, err := parseHostPort(binding.HostPort)
			if err != nil {
				return nil, fmt.Errorf("container port %s: %w", key, err)
			}
			if ephemeral {
				hostPort, err = candidateHostPort(ctx, protocol, forbidden, selected)
				if err != nil {
					return nil, fmt.Errorf("container port %s: choose host port: %w", key, err)
				}
			}
			if hostPort == 2375 || forbidden[hostPort] || selected[hostPort] {
				return nil, fmt.Errorf("container port %s: host port %d is reserved or already owned", key, hostPort)
			}
			selected[hostPort] = true
			plans = append(plans, portPlan{
				containerPort: key,
				ephemeral:     ephemeral,
				publication: portPublication{
					HostIP: "127.0.0.1", HostPort: hostPort, SandboxPort: hostPort, Protocol: protocol + "4",
				},
			})
		}
	}
	return plans, nil
}

func parseContainerPort(key string) (int, error) {
	raw, _, ok := strings.Cut(key, "/")
	if !ok || strings.ContainsAny(raw, "-:") {
		return 0, errors.New("invalid container port")
	}
	port, err := strconv.Atoi(raw)
	if err != nil || port < 1 || port > 65535 {
		return 0, errors.New("invalid container port")
	}
	return port, nil
}

func validateHostIP(raw string) error {
	switch raw {
	case "", "0.0.0.0", "127.0.0.1":
		return nil
	default:
		ip := net.ParseIP(raw)
		if ip != nil && ip.To4() == nil {
			return fmt.Errorf("IPv6 host binding %q is unsupported", raw)
		}
		return fmt.Errorf("host binding %q is not local loopback", raw)
	}
}

func parseHostPort(raw string) (int, bool, error) {
	if raw == "" || raw == "0" {
		return 0, true, nil
	}
	if strings.ContainsAny(raw, "-:") {
		return 0, false, fmt.Errorf("sbx requires one host port, got %q", raw)
	}
	port, err := strconv.Atoi(raw)
	if err != nil || port < 1 || port > 65535 {
		return 0, false, fmt.Errorf("invalid host port %q", raw)
	}
	return port, false, nil
}

func candidateHostPort(ctx context.Context, protocol string, forbidden, selected map[int]bool) (int, error) {
	for range portPublishTries * 2 {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		port, err := availableHostPort(protocol)
		if err != nil {
			return 0, err
		}
		if port != 2375 && !forbidden[port] && !selected[port] {
			return port, nil
		}
	}
	return 0, errors.New("could not find an unowned loopback port")
}

func availableHostPort(protocol string) (int, error) {
	if protocol == "udp" {
		conn, err := net.ListenPacket("udp4", "127.0.0.1:0")
		if err != nil {
			return 0, err
		}
		defer conn.Close()
		return conn.LocalAddr().(*net.UDPAddr).Port, nil
	}
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

func newLeaseToken() (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate port lease token: %w", err)
	}
	return hex.EncodeToString(raw[:]), nil
}

func (m *portManager) publish(ctx context.Context, token string, plan portPlan) (portPublication, bool, error) {
	want := plan.publication
	tried := map[int]bool{want.HostPort: true}
	tries := 1
	if plan.ephemeral {
		tries = portPublishTries
	}
	for attempt := range tries {
		if attempt > 0 {
			m.mu.Lock()
			forbidden := m.forbiddenPortsLocked()
			m.mu.Unlock()
			port, err := candidateHostPort(ctx, strings.TrimSuffix(want.Protocol, "4"), forbidden, tried)
			if err != nil {
				return portPublication{}, false, err
			}
			want.HostPort, want.SandboxPort = port, port
			tried[port] = true
		}
		commandCtx, cancel := boundedContext(ctx, portCommandTimeout)
		output, err := m.runtime.run(commandCtx, "ports", m.runtime.name, "--publish", want.spec(), "--format", "json")
		cancel()
		if err != nil {
			if plan.ephemeral && confirmedAddressInUse(err) {
				continue
			}
			return want, false, fmt.Errorf("sandbox %q creation %s port %s: publish with sbx: %w", m.runtime.name, token, want.spec(), err)
		}
		var rows []portPublication
		if decodeErr := json.Unmarshal(output, &rows); decodeErr != nil {
			return want, false, fmt.Errorf("sandbox %q creation %s port %s: decode sbx publish response %q: %w", m.runtime.name, token, want.spec(), string(output), decodeErr)
		}
		if len(rows) != 1 || rows[0] != want {
			return firstPublication(rows, want), len(rows) > 0, fmt.Errorf("sandbox %q creation %s port %s: sbx returned unexpected binding %q", m.runtime.name, token, want.spec(), string(output))
		}
		return rows[0], true, nil
	}
	return portPublication{}, false, fmt.Errorf("sandbox %q creation %s port %s: address remained in use after %d attempts", m.runtime.name, token, want.spec(), tries)
}

func firstPublication(rows []portPublication, fallback portPublication) portPublication {
	if len(rows) == 0 {
		return fallback
	}
	return rows[0]
}

func confirmedAddressInUse(err error) bool {
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "address already in use") || strings.Contains(text, "already in use") || strings.Contains(text, "port_publish_conflict")
}

func (m *portManager) unpublish(ctx context.Context, publication portPublication) error {
	commandCtx, cancel := boundedContext(ctx, portCommandTimeout)
	_, err := m.runtime.run(commandCtx, "ports", m.runtime.name, "--unpublish", publication.spec(), "--format", "json")
	cancel()
	if err != nil {
		return fmt.Errorf("sandbox %q port %s: unpublish with sbx: %w", m.runtime.name, publication.spec(), err)
	}
	published, err := m.listPublished(ctx)
	if err != nil {
		return fmt.Errorf("sandbox %q port %s: verify sbx unpublish: %w", m.runtime.name, publication.spec(), err)
	}
	if containsPublication(published, publication) {
		return fmt.Errorf("sandbox %q port %s: sbx reported successful unpublish but the binding remains published", m.runtime.name, publication.spec())
	}
	return nil
}

func boundedContext(parent context.Context, maximum time.Duration) (context.Context, context.CancelFunc) {
	if deadline, ok := parent.Deadline(); ok && time.Until(deadline) <= maximum {
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, maximum)
}
