package cockroachdb_test

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/cockroachdb"
)

func TestCockroach_Insecure(t *testing.T) {
	suite.Run(t, &AuthNSuite{
		url: "postgres://root@localhost:xxxxx/defaultdb?sslmode=disable",
	})
}

func TestCockroach_NotRoot(t *testing.T) {
	suite.Run(t, &AuthNSuite{
		url: "postgres://test@localhost:xxxxx/defaultdb?sslmode=disable",
		opts: []testcontainers.ContainerCustomizer{
			cockroachdb.WithUser("test"),
		},
	})
}

func TestCockroach_Password(t *testing.T) {
	suite.Run(t, &AuthNSuite{
		url: "postgres://foo:bar@localhost:xxxxx/defaultdb?sslmode=disable",
		opts: []testcontainers.ContainerCustomizer{
			cockroachdb.WithUser("foo"),
			cockroachdb.WithPassword("bar"),
		},
	})
}

func TestCockroach_TLS(t *testing.T) {
	tlsCfg, err := cockroachdb.NewTLSConfig()
	require.NoError(t, err)

	suite.Run(t, &AuthNSuite{
		url: "postgres://root@localhost:xxxxx/defaultdb?sslmode=verify-full",
		opts: []testcontainers.ContainerCustomizer{
			cockroachdb.WithTLS(tlsCfg),
		},
	})
}

type AuthNSuite struct {
	suite.Suite
	url  string
	opts []testcontainers.ContainerCustomizer
}

func (suite *AuthNSuite) TestConnectionString() {
	ctx := context.Background()

	container, err := cockroachdb.RunContainer(ctx, suite.opts...)
	suite.Require().NoError(err)

	suite.T().Cleanup(func() {
		err := container.Terminate(ctx)
		suite.Require().NoError(err)
	})

	connStr, err := removePort(container.MustConnectionString(ctx))
	suite.Require().NoError(err)

	suite.Equal(suite.url, connStr)
}

func (suite *AuthNSuite) TestPing() {
	ctx := context.Background()

	inputs := []struct {
		name string
		opts []testcontainers.ContainerCustomizer
	}{
		{
			name: "defaults",
			// opts: suite.opts
		},
		{
			name: "database",
			opts: []testcontainers.ContainerCustomizer{
				cockroachdb.WithDatabase("test"),
			},
		},
	}

	for _, input := range inputs {
		suite.Run(input.name, func() {
			opts := suite.opts
			opts = append(opts, input.opts...)

			container, err := cockroachdb.RunContainer(ctx, opts...)
			suite.Require().NoError(err)

			suite.T().Cleanup(func() {
				err := container.Terminate(ctx)
				suite.Require().NoError(err)
			})

			conn, err := conn(ctx, container)
			suite.Require().NoError(err)
			defer conn.Close(ctx)

			err = conn.Ping(ctx)
			suite.Require().NoError(err)
		})
	}
}

func (suite *AuthNSuite) TestQuery() {
	ctx := context.Background()

	container, err := cockroachdb.RunContainer(ctx, suite.opts...)
	suite.Require().NoError(err)

	suite.T().Cleanup(func() {
		err := container.Terminate(ctx)
		suite.Require().NoError(err)
	})

	conn, err := conn(ctx, container)
	suite.Require().NoError(err)
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx, "CREATE TABLE test (id INT PRIMARY KEY)")
	suite.Require().NoError(err)

	_, err = conn.Exec(ctx, "INSERT INTO test (id) VALUES (523123)")
	suite.Require().NoError(err)

	var id int
	err = conn.QueryRow(ctx, "SELECT id FROM test").Scan(&id)
	suite.Require().NoError(err)
	suite.Equal(523123, id)
}

func conn(ctx context.Context, container *cockroachdb.CockroachDBContainer) (*pgx.Conn, error) {
	cfg, err := pgx.ParseConfig(container.MustConnectionString(ctx))
	if err != nil {
		return nil, err
	}

	tlsCfg, err := container.TLSConfig()
	switch {
	case err != nil:
		if !errors.Is(err, cockroachdb.ErrTLSNotEnabled) {
			return nil, err
		}
	default:
		// apply TLS config
		cfg.TLSConfig = tlsCfg
	}

	return pgx.ConnectConfig(ctx, cfg)
}

func removePort(s string) (string, error) {
	u, err := url.Parse(s)
	if err != nil {
		return "", err
	}
	return strings.Replace(s, ":"+u.Port(), ":xxxxx", 1), nil
}
