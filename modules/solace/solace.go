package solace

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type SolaceContainer struct {
	testcontainers.Container
	username     string
	password     string
	vpn          string
	queues       map[string][]string // queueName -> topics
	image        string
	exposedPorts []string
	envVars      map[string]string
	shmSize      int64
}

// WithCredentials sets the client credentials (username, password)
func (s *SolaceContainer) WithCredentials(username, password string) *SolaceContainer {
	s.username = username
	s.password = password
	return s
}

// WithVpn sets the VPN name
func (s *SolaceContainer) WithVpn(vpn string) *SolaceContainer {
	s.vpn = vpn
	return s
}

// WithQueue subscribes a given topic to a queue (for SMF/AMQP testing)
func (s *SolaceContainer) WithQueue(queueName, topic string) *SolaceContainer {
	// Store queue/topic mapping in a new field for later use (e.g., CLI script generation)
	if s.queues == nil {
		s.queues = make(map[string][]string)
	}
	s.queues[queueName] = append(s.queues[queueName], topic)
	return s
}

// WithEnv allows adding or overriding environment variables
func (s *SolaceContainer) WithEnv(env map[string]string) *SolaceContainer {
	if s.envVars == nil {
		s.envVars = make(map[string]string)
	}
	for k, v := range env {
		s.envVars[k] = v
	}
	return s
}

// WithExposedPorts allows adding extra exposed ports
func (s *SolaceContainer) WithExposedPorts(ports ...string) *SolaceContainer {
	s.exposedPorts = append(s.exposedPorts, ports...)
	return s
}

const (
	DefaultImage = "solace/solace-pubsub-standard:latest"
	DefaultVpn   = "default"
	DefaultUser  = "root"
	DefaultPass  = "password"
)

func NewSolaceContainer(ctx context.Context, image string) *SolaceContainer {
	if image == "" {
		image = DefaultImage
	}
	return &SolaceContainer{
		username: DefaultUser,
		password: DefaultPass,
		vpn:      DefaultVpn,
		image:    image,
	}
}

func (s *SolaceContainer) WithShmSize(size int64) *SolaceContainer {
	s.shmSize = size
	return s
}

// waitForSolaceActive waits for the Solace broker to be fully active by checking the system log
func (s *SolaceContainer) waitForSolaceActive(ctx context.Context) error {
	const maxAttempts = 60
	const intervalSeconds = 1

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Execute grep command to check for "Primary Virtual Router is now active" in system log
		code, _, err := s.Exec(ctx, []string{"grep", "-R", "Primary Virtual Router is now active", "/usr/sw/jail/logs/system.log"})

		if err == nil && code == 0 {
			// Found the message, broker is active
			return nil
		}

		// Wait before next attempt
		time.Sleep(intervalSeconds * time.Second)
	}

	return fmt.Errorf("timeout waiting for Solace broker to become active after %d seconds", maxAttempts)
}

func (s *SolaceContainer) Run(ctx context.Context) error {
	shmSize := s.shmSize
	if shmSize == 0 {
		shmSize = 1 << 30 // Default 1GB
	}

	// Merge env vars
	env := map[string]string{
		"username_admin_globalaccesslevel": "admin",
		"username_admin_password":          "admin",
	}
	for k, v := range s.envVars {
		env[k] = v
	}

	// Collect exposed ports from all known services
	defaultServices := []Service{ServiceAMQP, ServiceManagement, ServiceSMF, ServiceREST, ServiceMQTT}
	portsMap := make(map[string]struct{})
	var ports []string
	for _, svc := range defaultServices {
		portStr := fmt.Sprintf("%d/tcp", svc.Port)
		if _, exists := portsMap[portStr]; !exists {
			ports = append(ports, portStr)
			portsMap[portStr] = struct{}{}
		}
	}
	if len(s.exposedPorts) > 0 {
		ports = append(ports, s.exposedPorts...)
	}

	// --- CLI script generation for queue/topic config ---
	// Reference: https://docs.solace.com/Admin-Ref/CLI-Reference/VMR_CLI_Commands.html
	var cliScript string
	if len(s.queues) > 0 {
		cliScript += "enable\nconfigure\n"

		// Create VPN if not default
		if s.vpn != DefaultVpn {
			cliScript += fmt.Sprintf("create message-vpn %s\n", s.vpn)
			cliScript += "no shutdown\n"
			cliScript += "exit\n"
			cliScript += fmt.Sprintf("client-profile default message-vpn %s\n", s.vpn)
			cliScript += "message-spool\n"
			cliScript += "allow-guaranteed-message-send\n"
			cliScript += "allow-guaranteed-message-receive\n"
			cliScript += "allow-guaranteed-endpoint-create\n"
			cliScript += "allow-guaranteed-endpoint-create-durability all\n"
			cliScript += "exit\n"
			cliScript += "exit\n"
			cliScript += fmt.Sprintf("message-spool message-vpn %s\n", s.vpn)
			cliScript += "max-spool-usage 60000\n"
			cliScript += "exit\n"
		}

		// Configure username and password
		cliScript += fmt.Sprintf("create client-username %s message-vpn %s\n", s.username, s.vpn)
		cliScript += fmt.Sprintf("password %s\n", s.password)
		cliScript += "no shutdown\n"
		cliScript += "exit\n"

		// Configure VPN Basic authentication
		cliScript += fmt.Sprintf("message-vpn %s\n", s.vpn)
		cliScript += "authentication basic auth-type internal\n"
		cliScript += "no shutdown\n"
		cliScript += "end\n"

		// Configure queues and topic subscriptions
		cliScript += "configure\n"
		cliScript += fmt.Sprintf("message-spool message-vpn %s\n", s.vpn)
		for queue, topics := range s.queues {
			// Create the queue first
			cliScript += fmt.Sprintf("create queue %s\n", queue)
			cliScript += "access-type exclusive\n"
			cliScript += "permission all consume\n"
			cliScript += "no shutdown\n"
			cliScript += "exit\n"

			// Add topic subscriptions to the queue
			for _, topic := range topics {
				cliScript += fmt.Sprintf("queue %s\n", queue)
				cliScript += fmt.Sprintf("subscription topic %s\n", topic)
				cliScript += "exit\n"
			}
		}
		cliScript += "exit\nexit\n"
	}

	req := testcontainers.ContainerRequest{
		Image:        s.image,
		ExposedPorts: ports,
		Env:          env,
		Cmd:          nil,
		WaitingFor:   wait.ForHTTP("/").WithPort("8080/tcp").WithStartupTimeout(1 * time.Minute),
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.ShmSize = shmSize
		},
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return err
	}

	s.Container = container // assign before using s.Container

	// Wait for Solace broker to be fully active before configuring
	if err := s.waitForSolaceActive(ctx); err != nil {
		return fmt.Errorf("failed waiting for Solace broker to become active: %w", err)
	}

	// Copy and execute CLI script inside the container if it was generated
	if cliScript != "" {
		// Write the CLI script to a temp file
		tmpFile, err := os.CreateTemp("", "solace-queue-setup-*.cli")
		if err != nil {
			return fmt.Errorf("failed to create temp CLI script: %w", err)
		}
		defer os.Remove(tmpFile.Name())
		if _, err := tmpFile.Write([]byte(cliScript)); err != nil {
			return fmt.Errorf("failed to write CLI script: %w", err)
		}
		if err := tmpFile.Close(); err != nil {
			return fmt.Errorf("failed to close CLI script: %w", err)
		}

		// Copy the script into the container at the correct location
		err = s.CopyFileToContainer(ctx, tmpFile.Name(), "/usr/sw/jail/cliscripts/script.cli", 0o644)
		if err != nil {
			return fmt.Errorf("failed to copy CLI script to container: %w", err)
		}

		// Execute the script
		code, out, err := s.Exec(ctx, []string{"/usr/sw/loads/currentload/bin/cli", "-A", "-es", "script.cli"})
		output := ""
		if out != nil {
			bytes, readErr := io.ReadAll(out)
			if readErr == nil {
				output = string(bytes)
			} else {
				output = fmt.Sprintf("[ERROR reading CLI output: %v]", readErr)
			}
		}
		if err != nil {
			return fmt.Errorf("failed to execute CLI script for queue/topic setup: %w", err)
		}
		if code != 0 {
			return fmt.Errorf("CLI script execution failed with exit code %d: %s", code, output)
		}
	}

	return nil
}

// BrokerURLFor returns the origin URL for a given service
func (s *SolaceContainer) BrokerURLFor(ctx context.Context, service Service) (string, error) {
	p := nat.Port(fmt.Sprintf("%d/tcp", service.Port))
	return s.PortEndpoint(ctx, p, service.Protocol)
}

func (s *SolaceContainer) Username() string {
	return s.username
}

func (s *SolaceContainer) Password() string {
	return s.password
}

func (s *SolaceContainer) Vpn() string {
	return s.vpn
}
