package dolt_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	// Import mysql into the scope of this package (required)
	_ "github.com/go-sql-driver/mysql"

	"github.com/testcontainers/testcontainers-go/modules/dolt"
)

func TestDolt(t *testing.T) {
	ctx := context.Background()

	container, err := dolt.RunContainer(ctx)
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
	// connectionString {
	connectionString, err := container.ConnectionString(ctx)
	// }
	if err != nil {
		t.Fatal(err)
	}

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		t.Errorf("error pinging db: %+v\n", err)
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS a_table ( \n" +
		" `col_1` VARCHAR(128) NOT NULL, \n" +
		" `col_2` VARCHAR(128) NOT NULL, \n" +
		" PRIMARY KEY (`col_1`, `col_2`) \n" +
		")")
	if err != nil {
		t.Errorf("error creating table: %+v\n", err)
	}
}

func TestDoltWithNonRootUserAndEmptyPassword(t *testing.T) {
	ctx := context.Background()

	_, err := dolt.RunContainer(ctx,
		dolt.WithDatabase("foo"),
		dolt.WithUsername("test"),
		dolt.WithPassword(""))
	if err.Error() != "empty password can be used only with the root user" {
		t.Fatal(err)
	}
}

func TestDoltWithPublicRemoteCloneUrl(t *testing.T) {
	ctx := context.Background()

	_, err := dolt.RunContainer(ctx,
		dolt.WithDatabase("foo"),
		dolt.WithUsername("test"),
		dolt.WithPassword("test"),
		dolt.WithScripts(filepath.Join("testdata", "check_clone_public.sh")),
		dolt.WithDoltCloneRemoteUrl("fake-remote-url"))
	if err != nil {
		t.Fatal(err)
	}
}

func createTestCredsFile(t *testing.T) string {
	file, err := os.CreateTemp(os.TempDir(), "prefix")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	_, err = file.WriteString("some-fake-creds")
	if err != nil {
		t.Fatal(err)
	}
	return file.Name()
}

func TestDoltWithPrivateRemoteCloneUrl(t *testing.T) {
	ctx := context.Background()

	filename := createTestCredsFile(t)
	defer os.RemoveAll(filename)
	_, err := dolt.RunContainer(ctx,
		dolt.WithDatabase("foo"),
		dolt.WithUsername("test"),
		dolt.WithPassword("test"),
		dolt.WithScripts(filepath.Join("testdata", "check_clone_private.sh")),
		dolt.WithDoltCloneRemoteUrl("fake-remote-url"),
		dolt.WithDoltCredsPublicKey("fake-public-key"),
		dolt.WithCredsFile(filename))
	if err != nil {
		t.Fatal(err)
	}
}

func TestDoltWithRootUserAndEmptyPassword(t *testing.T) {
	ctx := context.Background()

	container, err := dolt.RunContainer(ctx,
		dolt.WithDatabase("foo"),
		dolt.WithUsername("root"),
		dolt.WithPassword(""))
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
	connectionString := container.MustConnectionString(ctx)

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		t.Errorf("error pinging db: %+v\n", err)
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS a_table ( \n" +
		" `col_1` VARCHAR(128) NOT NULL, \n" +
		" `col_2` VARCHAR(128) NOT NULL, \n" +
		" PRIMARY KEY (`col_1`, `col_2`) \n" +
		")")
	if err != nil {
		t.Errorf("error creating table: %+v\n", err)
	}
}

func TestDoltWithScripts(t *testing.T) {
	ctx := context.Background()

	container, err := dolt.RunContainer(ctx,
		dolt.WithScripts(filepath.Join("testdata", "schema.sql")))
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
	connectionString := container.MustConnectionString(ctx)

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		t.Errorf("error pinging db: %+v\n", err)
	}
	stmt, err := db.Prepare("SELECT name from profile")
	if err != nil {
		t.Fatal(err)
	}
	defer stmt.Close()
	row := stmt.QueryRow()
	var name string
	err = row.Scan(&name)
	if err != nil {
		t.Errorf("error fetching data")
	}
	if name != "profile 1" {
		t.Fatal("The expected record was not found in the database.")
	}
}
