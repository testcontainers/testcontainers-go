package testcontainers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"
	"golang.org/x/sync/errgroup"

	"github.com/testcontainers/testcontainers-go/internal/core/network"
)

const (
	image string = "testcontainers/sshd:1.1.0"
	// hostInternal is the internal hostname used to reach the host from the container,
	// using the SSHD container as a bridge.
	hostInternal string = "host.testcontainers.internal"
	user         string = "root"
	sshPort             = "22"
)

// sshPassword is a random password generated for the SSHD container.
var sshPassword = uuid.NewString()

// exposeHostPorts performs all the necessary steps to expose the host ports to the container, leveraging
// the SSHD container to create the tunnel, and the container lifecycle hooks to manage the tunnel lifecycle.
// At least one port must be provided to expose.
// The steps are:
// 1. Create a new SSHD container.
// 2. Expose the host ports to the container after the container is ready.
// 3. Close the SSH sessions before killing the container.
func exposeHostPorts(ctx context.Context, req *ContainerRequest, p ...int) (ContainerLifecycleHooks, error) {
	var sshdConnectHook ContainerLifecycleHooks

	if len(p) == 0 {
		return sshdConnectHook, fmt.Errorf("no ports to expose")
	}

	// Use the first network of the container to connect to the SSHD container.
	var sshdFirstNetwork string
	if len(req.Networks) > 0 {
		sshdFirstNetwork = req.Networks[0]
	} else if sshdFirstNetwork == "bridge" && len(req.Networks) > 1 {
		sshdFirstNetwork = req.Networks[1]
	}

	opts := []ContainerCustomizer{}
	if len(req.Networks) > 0 {
		// get the first network of the container to connect the SSHD container to it.
		nw, err := network.Get(ctx, sshdFirstNetwork)
		if err != nil {
			return sshdConnectHook, fmt.Errorf("failed to get the network: %w", err)
		}

		dockerNw := DockerNetwork{
			ID:   nw.ID,
			Name: nw.Name,
		}

		// WithNetwork reuses an already existing network, attaching the container to it.
		// Finally it sets the network alias on that network to the given alias.
		// TODO: Using an anonymous function to avoid cyclic dependencies with the network package.
		withNetwork := func(aliases []string, nw *DockerNetwork) CustomizeRequestOption {
			return func(req *GenericContainerRequest) {
				networkName := nw.Name

				// attaching to the network because it was created with success or it already existed.
				req.Networks = append(req.Networks, networkName)

				if req.NetworkAliases == nil {
					req.NetworkAliases = make(map[string][]string)
				}
				req.NetworkAliases[networkName] = aliases
			}
		}

		opts = append(opts, withNetwork([]string{hostInternal}, &dockerNw))
	}

	// start the SSHD container with the provided options
	sshdContainer, err := newSshdContainer(ctx, opts...)
	if err != nil {
		return sshdConnectHook, fmt.Errorf("failed to create the SSH server: %w", err)
	}

	// IP in the first network of the container
	sshdIP, err := sshdContainer.ContainerIP(context.Background())
	if err != nil {
		return sshdConnectHook, fmt.Errorf("failed to get IP for the SSHD container: %w", err)
	}

	if req.HostConfigModifier == nil {
		req.HostConfigModifier = func(hostConfig *container.HostConfig) {}
	}

	// do not override the original HostConfigModifier
	originalHCM := req.HostConfigModifier
	req.HostConfigModifier = func(hostConfig *container.HostConfig) {
		// adding the host internal alias to the container as an extra host
		// to allow the container to reach the SSHD container.
		hostConfig.ExtraHosts = append(hostConfig.ExtraHosts, fmt.Sprintf("%s:%s", hostInternal, sshdIP))

		modes := []container.NetworkMode{container.NetworkMode(sshdFirstNetwork), "none", "host"}
		// if the container is not in one of the modes, attach it to the first network of the SSHD container
		found := false
		for _, mode := range modes {
			if hostConfig.NetworkMode == mode {
				found = true
				break
			}
		}
		if !found {
			req.Networks = append(req.Networks, sshdFirstNetwork)
		}

		// invoke the original HostConfigModifier with the updated hostConfig
		originalHCM(hostConfig)
	}

	// after the container is ready, create the SSH tunnel
	// for each exposed port from the host. We are going to
	// use an error group to expose all ports in parallel,
	// and return an error if any of them fails.
	sshdConnectHook = ContainerLifecycleHooks{
		PostReadies: []ContainerHook{
			func(ctx context.Context, c Container) error {
				var errs []error

				for _, exposedHostPort := range req.HostAccessPorts {
					err := sshdContainer.exposeHostPort(ctx, exposedHostPort)
					if err != nil {
						errs = append(errs, err)
					}
				}

				if len(errs) > 0 {
					return fmt.Errorf("failed to expose host ports: %w", errors.Join(errs...))
				}

				return nil
			},
		},
		PreTerminates: []ContainerHook{
			func(ctx context.Context, _ Container) error {
				// before killing the container, close the SSH sessions
				return sshdContainer.Terminate(ctx)
			},
		},
	}

	return sshdConnectHook, nil
}

// newSshdContainer creates a new SSHD container with the provided options.
func newSshdContainer(ctx context.Context, opts ...ContainerCustomizer) (*sshdContainer, error) {
	// Disable ipv6 & Make it listen on all interfaces, not just localhost
	// Enable algorithms supported by our ssh client library
	cmd := `echo "` + user + `:${PASSWORD}" | chpasswd && /usr/sbin/sshd -D -o PermitRootLogin=yes ` +
		`-o AddressFamily=inet -o GatewayPorts=yes -o AllowAgentForwarding=yes -o AllowTcpForwarding=yes ` +
		`-o KexAlgorithms=+diffie-hellman-group1-sha1 -o HostkeyAlgorithms=+ssh-rsa `

	req := GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			Image:           image,
			HostAccessPorts: []int{}, // empty list because it does not need any port
			ExposedPorts:    []string{sshPort},
			Env:             map[string]string{"PASSWORD": sshPassword},
			Cmd:             []string{"sh", "-c", cmd},
		},
		Started: true,
	}

	for _, opt := range opts {
		opt.Customize(&req)
	}

	c, err := GenericContainer(ctx, req)
	if err != nil {
		return nil, err
	}

	// force a type assertion to return a concrete type,
	// because GenericContainer returns a Container interface.
	dc := c.(*DockerContainer)

	sshd := &sshdContainer{
		DockerContainer: dc,
		tunnels:         make(map[int]*sshTunnel),
	}

	err = sshd.configureSSHSession(ctx)
	if err != nil {
		// return the container and the error to the caller to handle it
		return sshd, err
	}

	return sshd, nil
}

// sshdContainer represents the SSHD container type used for the port forwarding container.
// It's an internal type that extends the DockerContainer type, to add the SSH tunneling capabilities.
type sshdContainer struct {
	*DockerContainer
	port      string
	sshConfig *ssh.ClientConfig
	tunnels   map[int]*sshTunnel
}

// Terminate stops the container and closes the SSH session
func (sshdC *sshdContainer) Terminate(ctx context.Context) error {
	for _, t := range sshdC.tunnels {
		defer t.Close()
	}

	return sshdC.DockerContainer.Terminate(ctx)
}

func (sshdC *sshdContainer) configureSSHSession(ctx context.Context) error {
	if sshdC.sshConfig != nil {
		// do not configure the SSH session twice
		return nil
	}

	mappedPort, err := sshdC.MappedPort(ctx, sshPort)
	if err != nil {
		return err
	}
	sshdC.port = mappedPort.Port()

	sshConfig := ssh.ClientConfig{
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth:            []ssh.AuthMethod{ssh.Password(sshPassword)},
		Timeout:         30 * time.Second,
	}

	sshdC.sshConfig = &sshConfig

	return nil
}

func (sshdC *sshdContainer) exposeHostPort(ctx context.Context, port int) error {
	if _, ok := sshdC.tunnels[port]; ok {
		// do not expose the same port twice
		return nil
	}

	// Setup the tunnel, but do not yet start it yet.
	tunnel := newSSHTunnel(
		fmt.Sprintf("%s@localhost:%s", user, sshdC.port),
		sshdC.sshConfig,
		"localhost", port, // The destination host and port of the actual server.
	)

	// use testcontainers logger
	tunnel.Log = Logger

	err := tunnel.Start(ctx)
	if err != nil {
		return fmt.Errorf("failed to start the SSH tunnel: %w", err)
	}

	sshdC.tunnels[port] = tunnel

	return nil
}

type sshEndpoint struct {
	Host string
	Port int
	User string
}

func newSshEndpoint(s string) *sshEndpoint {
	endpoint := &sshEndpoint{
		Host: s,
	}
	if parts := strings.Split(endpoint.Host, "@"); len(parts) > 1 {
		endpoint.User = parts[0]
		endpoint.Host = parts[1]
	}
	if parts := strings.Split(endpoint.Host, ":"); len(parts) > 1 {
		endpoint.Host = parts[0]
		endpoint.Port, _ = strconv.Atoi(parts[1])
	}
	return endpoint
}

func (endpoint *sshEndpoint) String() string {
	return fmt.Sprintf("%s:%d", endpoint.Host, endpoint.Port)
}

type sshTunnel struct {
	Local  *sshEndpoint
	Server *sshEndpoint
	Remote *sshEndpoint
	Config *ssh.ClientConfig
	Log    Logging
}

func newSSHTunnel(tunnel string, sshConfig *ssh.ClientConfig, target string, targetPort int) *sshTunnel {
	destination := fmt.Sprintf("%s:%d", target, targetPort)

	localEndpoint := newSshEndpoint(destination)

	server := newSshEndpoint(tunnel)
	if server.Port == 0 {
		server.Port = 22
	}

	return &sshTunnel{
		Config: sshConfig,
		Local:  localEndpoint,
		Server: server,
		Remote: newSshEndpoint(destination),
	}
}

func (tunnel *sshTunnel) logf(fmt string, args ...interface{}) {
	if tunnel.Log != nil {
		tunnel.Log.Printf(fmt, args...)
	}
}

func (tunnel *sshTunnel) Close() error {
	if tunnel.Log != nil {
		tunnel.logf("closing tunnel")
	}

	return nil
}

func (tunnel *sshTunnel) Start(ctx context.Context) error {
	lcfg := net.ListenConfig{}

	listener, err := lcfg.Listen(ctx, "tcp", tunnel.Local.String())
	if err != nil {
		return err
	}
	defer listener.Close()

	tunnel.Local.Port = listener.Addr().(*net.TCPAddr).Port
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		tunnel.logf("accepted connection")
		tunnel.forward(conn)
	}
}

func (tunnel *sshTunnel) forward(localConn net.Conn) error {
	serverConn, err := ssh.Dial("tcp", tunnel.Server.String(), tunnel.Config)
	if err != nil {
		return fmt.Errorf("server dial error: %s", err)
	}
	tunnel.logf("connected to %s (1 of 2)\n", tunnel.Server.String())

	remoteConn, err := serverConn.Dial("tcp", tunnel.Remote.String())
	if err != nil {
		return fmt.Errorf("remote dial error: %s", err)
	}
	tunnel.logf("connected to %s (2 of 2)\n", tunnel.Remote.String())

	copyConn := func(writer, reader net.Conn) error {
		_, err := io.Copy(writer, reader)
		if err != nil {
			return fmt.Errorf("io.Copy error: %s", err)
		}

		return nil
	}

	errgr := errgroup.Group{}

	errgr.Go(func() error {
		return copyConn(localConn, remoteConn)
	})
	errgr.Go(func() error {
		return copyConn(remoteConn, localConn)
	})

	return errgr.Wait()
}
