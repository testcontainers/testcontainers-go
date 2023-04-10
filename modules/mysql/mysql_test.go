package mysql

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	// Import mysql into the scope of this package (required)
	_ "github.com/go-sql-driver/mysql"
	"github.com/testcontainers/testcontainers-go"
)

func TestMySQL(t *testing.T) {
	ctx := context.Background()

	// createMysqlContainer {
	container, err := RunContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}
	// }

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// perform assertions
	connectionString, _ := container.ConnectionString(ctx, "tls=skip-verify")

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

func TestMySQLWithNonRootUserAndEmptyPassword(t *testing.T) {
	ctx := context.Background()

	_, err := RunContainer(ctx,
		WithDatabase("foo"),
		WithUsername("test"),
		WithPassword(""))
	if err.Error() != "empty password can be used only with the root user" {
		t.Fatal(err)
	}
}

func TestMySQLWithRootUserAndEmptyPassword(t *testing.T) {
	ctx := context.Background()

	// customInitialization {
	container, err := RunContainer(ctx,
		WithDatabase("foo"),
		WithUsername("root"),
		WithPassword(""))
	if err != nil {
		t.Fatal(err)
	}
	// }

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// perform assertions
	connectionString, _ := container.ConnectionString(ctx)

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

func TestMySQLWithConfigFile(t *testing.T) {
	ctx := context.Background()

	// withConfigFile {
	container, err := RunContainer(ctx, testcontainers.WithImage("mysql:5.6"),
		WithConfigFile("./testdata/my.cnf"))
	if err != nil {
		t.Fatal(err)
	}
	// }

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// perform assertions
	connectionString, _ := container.ConnectionString(ctx)

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		t.Errorf("error pinging db: %+v\n", err)
	}
	stmt, _ := db.Prepare("SELECT @@GLOBAL.innodb_file_format")
	if err != nil {
		t.Fatal(err)
	}
	defer stmt.Close()
	row := stmt.QueryRow()
	var innodbFileFormat = ""
	err = row.Scan(&innodbFileFormat)
	if err != nil {
		t.Errorf("error fetching innodb_file_format value")
	}
	if innodbFileFormat != "Barracuda" {
		t.Fatal("The InnoDB file format has been set by the ini file content")
	}
}

func TestMySQLWithScripts(t *testing.T) {
	ctx := context.Background()

	// withScripts {
	container, err := RunContainer(ctx,
		WithScripts(filepath.Join("testdata", "schema.sql")))
	if err != nil {
		t.Fatal(err)
	}
	// }

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// perform assertions
	connectionString, _ := container.ConnectionString(ctx)

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
