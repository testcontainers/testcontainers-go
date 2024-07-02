package mssql_test

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/microsoft/go-mssqldb"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mssql"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestMSSQLServer(t *testing.T) {
	ctx := context.Background()

	container, err := mssql.Run(ctx,
		"mcr.microsoft.com/mssql/server:2022-CU10-ubuntu-22.04",
		mssql.WithAcceptEULA(),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// perform assertions
	connectionString, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatal(err)
	}

	db, err := sql.Open("sqlserver", connectionString)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		t.Errorf("error pinging db: %+v\n", err)
	}

	_, err = db.Exec("CREATE TABLE a_table ( " +
		" [col_1] NVARCHAR(128) NOT NULL, " +
		" [col_2] NVARCHAR(128) NOT NULL, " +
		" PRIMARY KEY ([col_1], [col_2]) " +
		")")
	if err != nil {
		t.Errorf("error creating table: %+v\n", err)
	}
}

func TestMSSQLServerWithMissingEulaOption(t *testing.T) {
	ctx := context.Background()

	container, err := mssql.Run(ctx,
		"mcr.microsoft.com/mssql/server:2022-CU10-ubuntu-22.04",
		testcontainers.WithWaitStrategy(
			wait.ForLog("The SQL Server End-User License Agreement (EULA) must be accepted")),
	)
	if err != nil {
		t.Fatalf("Expected a log to confirm missing EULA but got error: %s", err)
	}

	state, err := container.State(ctx)
	if err != nil {
		t.Fatalf("failed to get container state: %s", err)
	}

	if !state.Running {
		t.Log("Success: Confirmed proper handling of missing EULA, so container is not running.")
	}
}

func TestMSSQLServerWithConnectionStringParameters(t *testing.T) {
	ctx := context.Background()

	container, err := mssql.Run(ctx,
		"mcr.microsoft.com/mssql/server:2022-CU10-ubuntu-22.04",
		mssql.WithAcceptEULA(),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// perform assertions
	connectionString, err := container.ConnectionString(ctx, "encrypt=false", "TrustServerCertificate=true")
	if err != nil {
		t.Fatal(err)
	}

	db, err := sql.Open("sqlserver", connectionString)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		t.Errorf("error pinging db: %+v\n", err)
	}

	_, err = db.Exec("CREATE TABLE a_table ( " +
		" [col_1] NVARCHAR(128) NOT NULL, " +
		" [col_2] NVARCHAR(128) NOT NULL, " +
		" PRIMARY KEY ([col_1], [col_2]) " +
		")")
	if err != nil {
		t.Errorf("error creating table: %+v\n", err)
	}
}

func TestMSSQLServerWithCustomStrongPassword(t *testing.T) {
	ctx := context.Background()

	container, err := mssql.Run(ctx,
		"mcr.microsoft.com/mssql/server:2022-CU10-ubuntu-22.04",
		mssql.WithAcceptEULA(),
		mssql.WithPassword("Strong@Passw0rd"),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// perform assertions
	connectionString, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatal(err)
	}

	db, err := sql.Open("sqlserver", connectionString)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		t.Errorf("error pinging db: %+v\n", err)
	}
}

// tests that a weak password is not accepted by the container due to Microsoft's password strength policy
func TestMSSQLServerWithInvalidPassword(t *testing.T) {
	ctx := context.Background()

	container, err := mssql.Run(ctx,
		"mcr.microsoft.com/mssql/server:2022-CU10-ubuntu-22.04",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Password validation failed")),
		mssql.WithAcceptEULA(),
		mssql.WithPassword("weakPassword"),
	)

	if err == nil {
		t.Log("Success: Received invalid password validation docker log.")
	} else {
		t.Fatalf("Expected a password validation log but got error: %s", err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})
}

func TestMSSQLServerWithAlternativeImage(t *testing.T) {
	ctx := context.Background()

	container, err := mssql.Run(ctx,
		"mcr.microsoft.com/mssql/server:2022-RTM-GDR1-ubuntu-20.04",
		mssql.WithAcceptEULA(),
	)
	if err != nil {
		t.Fatalf("Failed to create the container with alternative image: %s", err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// perform assertions
	connectionString, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatal(err)
	}

	db, err := sql.Open("sqlserver", connectionString)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		t.Errorf("error pinging db: %+v\n", err)
	}
}
