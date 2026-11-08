// Package dexexamples provides optional Dex gRPC administration adapters.
package dexexamples

import (
	"context"
	"errors"
	"fmt"

	"github.com/dexidp/dex/api/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/testcontainers/testcontainers-go/modules/dex"
)

// AdminClient administers a running Dex container via its plaintext gRPC API.
// Each operation opens and closes its own connection. Provisioning and static
// registration remain in the parent dex package. Not safe for concurrent use.
type AdminClient struct {
	container *dex.Container
}

// NewAdminClient creates an admin adapter. It does not open a connection.
func NewAdminClient(c *dex.Container) *AdminClient {
	return &AdminClient{container: c}
}

// AddClient creates a client via Dex's gRPC admin API.
//
// Note: Dex's api.Client proto has no grant_types field — clients added this
// way inherit Dex's defaults (authorization_code + refresh_token). For
// custom grants, register the client pre-start via WithClient.
//
// Not safe for concurrent use.
func (c *AdminClient) AddClient(ctx context.Context, cl dex.Client) error {
	target, err := c.container.GRPCEndpoint(ctx)
	if err != nil {
		return err
	}
	conn, err := dial(target)
	if err != nil {
		return err
	}
	defer conn.Close()

	registration := cl.Registration()
	resp, err := api.NewDexClient(conn).CreateClient(ctx, &api.CreateClientReq{
		Client: &api.Client{
			Id:           registration.ID,
			Secret:       registration.Secret,
			RedirectUris: registration.RedirectURIs,
			Name:         registration.Name,
			Public:       registration.Public,
		},
	})
	if err != nil {
		return fmt.Errorf("dex: create client: %w", err)
	}
	if resp.AlreadyExists {
		return fmt.Errorf("%w: %q", dex.ErrClientExists, registration.ID)
	}
	return nil
}

// RemoveClient deletes a client by ID.
//
// Not safe for concurrent use.
func (c *AdminClient) RemoveClient(ctx context.Context, id string) error {
	target, err := c.container.GRPCEndpoint(ctx)
	if err != nil {
		return err
	}
	conn, err := dial(target)
	if err != nil {
		return err
	}
	defer conn.Close()

	resp, err := api.NewDexClient(conn).DeleteClient(ctx, &api.DeleteClientReq{Id: id})
	if err != nil {
		return fmt.Errorf("dex: delete client: %w", err)
	}
	if resp.NotFound {
		return fmt.Errorf("%w: %q", dex.ErrClientNotFound, id)
	}
	return nil
}

// AddUser registers a user in Dex's password DB via gRPC.
//
// Not safe for concurrent use.
func (c *AdminClient) AddUser(ctx context.Context, u dex.User) error {
	target, err := c.container.GRPCEndpoint(ctx)
	if err != nil {
		return err
	}
	conn, err := dial(target)
	if err != nil {
		return err
	}
	defer conn.Close()

	registration, err := u.PasswordRegistration()
	if err != nil {
		return err
	}

	resp, err := api.NewDexClient(conn).CreatePassword(ctx, &api.CreatePasswordReq{
		Password: &api.Password{
			Email:    registration.Email,
			Hash:     registration.Hash,
			Username: registration.Username,
			UserId:   registration.UserID,
		},
	})
	if err != nil {
		return fmt.Errorf("dex: create password: %w", err)
	}
	if resp.AlreadyExists {
		return fmt.Errorf("%w: %q", dex.ErrUserExists, registration.Email)
	}
	return nil
}

// RemoveUser deletes a user by email.
//
// Not safe for concurrent use.
func (c *AdminClient) RemoveUser(ctx context.Context, email string) error {
	target, err := c.container.GRPCEndpoint(ctx)
	if err != nil {
		return err
	}
	conn, err := dial(target)
	if err != nil {
		return err
	}
	defer conn.Close()

	resp, err := api.NewDexClient(conn).DeletePassword(ctx, &api.DeletePasswordReq{Email: email})
	if err != nil {
		return fmt.Errorf("dex: delete password: %w", err)
	}
	if resp.NotFound {
		return fmt.Errorf("%w: %q", dex.ErrUserNotFound, email)
	}
	return nil
}

func dial(target string) (*grpc.ClientConn, error) {
	if target == "" {
		return nil, errors.New("dex: grpc endpoint empty (container not started)")
	}
	conn, err := grpc.NewClient(target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("dex: grpc dial %s: %w", target, err)
	}
	return conn, nil
}
