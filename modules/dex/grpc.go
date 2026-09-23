package dex

import (
	"context"
	"errors"
	"fmt"

	"github.com/dexidp/dex/api/v2"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// AddClient creates a client via Dex's gRPC admin API.
//
// Note: Dex's api.Client proto has no grant_types field — clients added this
// way inherit Dex's defaults (authorization_code + refresh_token). For
// custom grants, register the client pre-start via WithClient.
//
// Not safe for concurrent use.
func (c *Container) AddClient(ctx context.Context, cl Client) error {
	target, err := c.GRPCEndpoint(ctx)
	if err != nil {
		return err
	}
	conn, err := dial(target)
	if err != nil {
		return err
	}
	defer conn.Close()

	resp, err := api.NewDexClient(conn).CreateClient(ctx, &api.CreateClientReq{
		Client: &api.Client{
			Id:           cl.id,
			Secret:       cl.secret,
			RedirectUris: cl.redirectURIs,
			Name:         cl.name,
			Public:       cl.public,
		},
	})
	if err != nil {
		return fmt.Errorf("dex: create client: %w", err)
	}
	if resp.AlreadyExists {
		return fmt.Errorf("%w: %q", ErrClientExists, cl.id)
	}
	return nil
}

// RemoveClient deletes a client by ID.
//
// Not safe for concurrent use.
func (c *Container) RemoveClient(ctx context.Context, id string) error {
	target, err := c.GRPCEndpoint(ctx)
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
		return fmt.Errorf("%w: %q", ErrClientNotFound, id)
	}
	return nil
}

// AddUser registers a user in Dex's password DB via gRPC.
//
// Not safe for concurrent use.
func (c *Container) AddUser(ctx context.Context, u User) error {
	target, err := c.GRPCEndpoint(ctx)
	if err != nil {
		return err
	}
	conn, err := dial(target)
	if err != nil {
		return err
	}
	defer conn.Close()

	hash, err := bcrypt.GenerateFromPassword([]byte(u.password), testBcryptCost)
	if err != nil {
		return fmt.Errorf("dex: bcrypt: %w", err)
	}
	userID := u.userID
	if userID == "" {
		uid, uidErr := newUUIDv4()
		if uidErr != nil {
			return fmt.Errorf("dex: generate user id: %w", uidErr)
		}
		userID = uid
	}

	resp, err := api.NewDexClient(conn).CreatePassword(ctx, &api.CreatePasswordReq{
		Password: &api.Password{
			Email:    u.email,
			Hash:     hash,
			Username: u.username,
			UserId:   userID,
		},
	})
	if err != nil {
		return fmt.Errorf("dex: create password: %w", err)
	}
	if resp.AlreadyExists {
		return fmt.Errorf("%w: %q", ErrUserExists, u.email)
	}
	return nil
}

// RemoveUser deletes a user by email.
//
// Not safe for concurrent use.
func (c *Container) RemoveUser(ctx context.Context, email string) error {
	target, err := c.GRPCEndpoint(ctx)
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
		return fmt.Errorf("%w: %q", ErrUserNotFound, email)
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
