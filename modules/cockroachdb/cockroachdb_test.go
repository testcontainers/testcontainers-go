package cockroachdb_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/cockroachdb"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestCockroach_Insecure(t *testing.T) {
	suite.Run(t, &AuthNSuite{})
}

func TestCockroach_NotRoot(t *testing.T) {
	suite.Run(t, &AuthNSuite{
		opts: []testcontainers.ContainerCustomizer{
			cockroachdb.WithUser("test"),
		},
	})
}

func TestCockroach_Password(t *testing.T) {
	suite.Run(t, &AuthNSuite{
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
		opts: []testcontainers.ContainerCustomizer{
			tlsCfg,
		},
	})
}

// TODO: remove this tests, its just a simple example for speed up the development.
func TestTLS(t *testing.T) {
	tlsCfg, err := cockroachdb.NewTLSConfig()
	require.NoError(t, err)

	ctx := context.Background()

	ctr, err := cockroachdb.Run(ctx, "cockroachdb/cockroach:latest-v23.1", tlsCfg)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)
	require.NotNil(t, ctr)
}

func TestNoTLS(t *testing.T) {
	ctx := context.Background()
	ctr, err := cockroachdb.Run(ctx, "cockroachdb/cockroach:latest-v23.1")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)
	require.NotNil(t, ctr)
}

func TestUser(t *testing.T) {
	ctx := context.Background()
	ctr, err := cockroachdb.Run(ctx, "cockroachdb/cockroach:latest-v23.1", cockroachdb.WithUser("test"))
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)
	require.NotNil(t, ctr)
}

type AuthNSuite struct {
	suite.Suite
	opts []testcontainers.ContainerCustomizer
}

func (suite *AuthNSuite) TestConnectionString() {
	ctx := context.Background()

	ctr, err := cockroachdb.Run(ctx, "cockroachdb/cockroach:latest-v23.1", suite.opts...)
	testcontainers.CleanupContainer(suite.T(), ctr)
	suite.Require().NoError(err)
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

			ctr, err := cockroachdb.Run(ctx, "cockroachdb/cockroach:latest-v23.1", opts...)
			testcontainers.CleanupContainer(suite.T(), ctr)
			suite.Require().NoError(err)

			conn, err := conn(ctx, ctr)
			suite.Require().NoError(err)
			defer conn.Close(ctx)

			err = conn.Ping(ctx)
			suite.Require().NoError(err)
		})
	}
}

func (suite *AuthNSuite) TestQuery() {
	ctx := context.Background()

	ctr, err := cockroachdb.Run(ctx, "cockroachdb/cockroach:latest-v23.1", suite.opts...)
	testcontainers.CleanupContainer(suite.T(), ctr)
	suite.Require().NoError(err)

	conn, err := conn(ctx, ctr)
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

// TestWithWaitStrategyAndDeadline covers a previous regression, container creation needs to fail to cover that path.
func (suite *AuthNSuite) TestWithWaitStrategyAndDeadline() {
	nodeStartUpCompleted := "node startup completed"

	suite.Run("Expected Failure To Run", func() {
		ctx := context.Background()

		// This will never match a log statement
		suite.opts = append(suite.opts, testcontainers.WithWaitStrategyAndDeadline(time.Millisecond*250, wait.ForLog("Won't Exist In Logs")))
		ctr, err := cockroachdb.Run(ctx, "cockroachdb/cockroach:latest-v23.1", suite.opts...)
		testcontainers.CleanupContainer(suite.T(), ctr)
		suite.Require().ErrorIs(err, context.DeadlineExceeded)
	})

	suite.Run("Expected Failure To Run But Would Succeed ", func() {
		ctx := context.Background()

		// This will timeout as we didn't give enough time for intialization, but would have succeeded otherwise
		suite.opts = append(suite.opts, testcontainers.WithWaitStrategyAndDeadline(time.Millisecond*20, wait.ForLog(nodeStartUpCompleted)))
		ctr, err := cockroachdb.Run(ctx, "cockroachdb/cockroach:latest-v23.1", suite.opts...)
		testcontainers.CleanupContainer(suite.T(), ctr)
		suite.Require().ErrorIs(err, context.DeadlineExceeded)
	})

	suite.Run("Succeeds And Executes Commands", func() {
		ctx := context.Background()

		// This will succeed
		suite.opts = append(suite.opts, testcontainers.WithWaitStrategyAndDeadline(time.Second*60, wait.ForLog(nodeStartUpCompleted)))
		ctr, err := cockroachdb.Run(ctx, "cockroachdb/cockroach:latest-v23.1", suite.opts...)
		testcontainers.CleanupContainer(suite.T(), ctr)
		suite.Require().NoError(err)

		conn, err := conn(ctx, ctr)
		suite.Require().NoError(err)
		defer conn.Close(ctx)

		_, err = conn.Exec(ctx, "CREATE TABLE test (id INT PRIMARY KEY)")
		suite.Require().NoError(err)
	})

	suite.Run("Succeeds And Executes Commands Waiting on HTTP Endpoint", func() {
		ctx := context.Background()

		// This will succeed
		suite.opts = append(suite.opts, testcontainers.WithWaitStrategyAndDeadline(time.Second*60, wait.ForHTTP("/health").WithPort("8080/tcp")))
		ctr, err := cockroachdb.Run(ctx, "cockroachdb/cockroach:latest-v23.1", suite.opts...)
		testcontainers.CleanupContainer(suite.T(), ctr)
		suite.Require().NoError(err)

		conn, err := conn(ctx, ctr)
		suite.Require().NoError(err)
		defer conn.Close(ctx)

		_, err = conn.Exec(ctx, "CREATE TABLE test (id INT PRIMARY KEY)")
		suite.Require().NoError(err)
	})
}

func conn(ctx context.Context, container *cockroachdb.CockroachDBContainer) (*pgx.Conn, error) {
	cfg, err := container.ConnectionConfig(ctx)
	if err != nil {
		return nil, err
	}

	return pgx.ConnectConfig(ctx, cfg)
}
