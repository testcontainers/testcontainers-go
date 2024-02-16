package mariadb_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	// Import mysql into the scope of this package (required)
	_ "github.com/go-sql-driver/mysql"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mariadb"
)

func TestMariaDB(t *testing.T) {
	ctx := context.Background()

	container, err := mariadb.RunContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// connectionString {
	// By default, MariaDB transmits data between the server and clients without encrypting it.
	connectionString, err := container.ConnectionString(ctx, "tls=false")
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

func TestMariaDBWithNonRootUserAndEmptyPassword(t *testing.T) {
	ctx := context.Background()

	_, err := mariadb.RunContainer(ctx,
		mariadb.WithDatabase("foo"),
		mariadb.WithUsername("test"),
		mariadb.WithPassword(""))
	if err.Error() != "empty password can be used only with the root user" {
		t.Fatal(err)
	}
}

func TestMariaDBWithRootUserAndEmptyPassword(t *testing.T) {
	ctx := context.Background()

	container, err := mariadb.RunContainer(ctx,
		mariadb.WithDatabase("foo"),
		mariadb.WithUsername("root"),
		mariadb.WithPassword(""))
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	connectionString, err := container.ConnectionString(ctx)
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

func TestMariaDBWithMySQLEnvVars(t *testing.T) {
	ctx := context.Background()

	container, err := mariadb.RunContainer(ctx, testcontainers.WithImage("mariadb:10.3.29"),
		mariadb.WithScripts(filepath.Join("testdata", "schema.sql")))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	assertDataCanBeFetched(t, ctx, container)
}

func TestMariaDBWithConfigFile(t *testing.T) {
	ctx := context.Background()

	container, err := mariadb.RunContainer(ctx, testcontainers.WithImage("mariadb:11.0.3"),
		mariadb.WithConfigFile(filepath.Join("testdata", "my.cnf")))
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	connectionString, err := container.ConnectionString(ctx)
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

	// In MariaDB 10.2.2 and later, the default file format is Barracuda and Antelope is deprecated.
	// Barracuda is a newer InnoDB file format. It supports the COMPACT, REDUNDANT, DYNAMIC and
	// COMPRESSED row formats. Tables with large BLOB or TEXT columns in particular could benefit
	// from the dynamic row format.
	stmt, err := db.Prepare("SELECT @@GLOBAL.innodb_default_row_format")
	if err != nil {
		t.Fatal(err)
	}
	defer stmt.Close()
	row := stmt.QueryRow()
	innodbFileFormat := ""
	err = row.Scan(&innodbFileFormat)
	if err != nil {
		t.Errorf("error fetching innodb_default_row_format value")
	}
	if innodbFileFormat != "dynamic" {
		t.Fatal("The InnoDB file format has been set by the ini file content")
	}
}

func TestMariaDBWithScripts(t *testing.T) {
	ctx := context.Background()

	container, err := mariadb.RunContainer(ctx,
		mariadb.WithScripts(filepath.Join("testdata", "schema.sql")))
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	assertDataCanBeFetched(t, ctx, container)
}

func assertDataCanBeFetched(t *testing.T, ctx context.Context, container *mariadb.MariaDBContainer) {
	connectionString, err := container.ConnectionString(ctx)
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
