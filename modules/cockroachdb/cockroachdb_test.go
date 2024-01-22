package cockroachdb_test

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/suite"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/cockroachdb"
)

func TestCockroach_Insecure(t *testing.T) {
	suite.Run(t, &CockroachDBSuite{
		url: "postgres://root@localhost:xxxxx/defaultdb?sslmode=disable",
	})
}

func TestCockroach_NonRootAuthn(t *testing.T) {
	suite.Run(t, &CockroachDBSuite{
		url: "postgres://test@localhost:xxxxx/defaultdb?sslmode=disable",
		opts: []testcontainers.ContainerCustomizer{
			cockroachdb.WithUser("test"),
		},
	})
}

func TestCockroach_PasswordAuthn(t *testing.T) {
	suite.Run(t, &CockroachDBSuite{
		url: "postgres://foo:bar@localhost:xxxxx/defaultdb?sslmode=disable",
		opts: []testcontainers.ContainerCustomizer{
			cockroachdb.WithUser("foo"),
			cockroachdb.WithPassword("bar"),
		},
	})
}

type CockroachDBSuite struct {
	suite.Suite
	url  string
	opts []testcontainers.ContainerCustomizer
}

func (suite *CockroachDBSuite) TestConnectionString() {
	ctx := context.Background()

	container, err := cockroachdb.RunContainer(ctx, suite.opts...)
	suite.NoError(err)

	suite.T().Cleanup(func() {
		err := container.Terminate(ctx)
		suite.NoError(err)
	})

	connStr, err := removePort(container.MustConnectionString(ctx))
	suite.NoError(err)

	suite.Equal(suite.url, connStr)
}

func (suite *CockroachDBSuite) TestPing() {
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
			opts := append(suite.opts, input.opts...)
			container, err := cockroachdb.RunContainer(ctx, opts...)
			suite.NoError(err)

			suite.T().Cleanup(func() {
				err := container.Terminate(ctx)
				suite.NoError(err)
			})

			conn, err := pgx.Connect(ctx, container.MustConnectionString(ctx))
			suite.NoError(err)

			err = conn.Ping(ctx)
			suite.NoError(err)
		})
	}
}

func (suite *CockroachDBSuite) TestQuery() {
	ctx := context.Background()

	container, err := cockroachdb.RunContainer(ctx, suite.opts...)
	suite.NoError(err)

	suite.T().Cleanup(func() {
		err := container.Terminate(ctx)
		suite.NoError(err)
	})

	conn, err := pgx.Connect(ctx, container.MustConnectionString(ctx))
	suite.NoError(err)

	_, err = conn.Exec(ctx, "CREATE TABLE test (id INT PRIMARY KEY)")
	suite.NoError(err)

	_, err = conn.Exec(ctx, "INSERT INTO test (id) VALUES (523123)")
	suite.NoError(err)

	var id int
	err = conn.QueryRow(ctx, "SELECT id FROM test").Scan(&id)
	suite.NoError(err)
	suite.Equal(523123, id)
}

func removePort(s string) (string, error) {
	u, err := url.Parse(s)
	if err != nil {
		return "", err
	}
	return strings.Replace(s, ":"+u.Port(), ":xxxxx", 1), nil
}
