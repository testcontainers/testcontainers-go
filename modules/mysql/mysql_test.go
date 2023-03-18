package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	// Import mysql into the scope of this package (required)
	_ "github.com/go-sql-driver/mysql"
)

func TestMySQL(t *testing.T) {
	ctx := context.Background()

	container, err := StartContainer(ctx, "mysql:8")
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

	host, _ := container.Host(ctx)

	p, _ := container.MappedPort(ctx, "3306/tcp")
	port := p.Int()

	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?tls=skip-verify",
		container.Username(), container.Password(), host, port, container.Database())

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

	_, err := StartContainer(ctx, "mysql:8", WithDatabase("foo"), WithUsername("test"), WithPassword(""))
	if err.Error() != "empty password can be used only with the root user" {
		t.Fatal(err)
	}
}

func TestMySQLWithRootUserAndEmptyPassword(t *testing.T) {
	ctx := context.Background()

	container, err := StartContainer(ctx, "mysql:8", WithDatabase("foo"), WithUsername("root"), WithPassword(""))
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

	host, _ := container.Host(ctx)

	p, _ := container.MappedPort(ctx, "3306/tcp")
	port := p.Int()

	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?tls=skip-verify",
		container.Username(), container.Password(), host, port, container.Database())

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

//func TestMySQLWithConfigFile(t *testing.T) {
//	ctx := context.Background()
//
//	container, err := StartContainer(ctx, "mysql:5.6.51", WithConfigFile("./conf.d/my.cnf"))
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	// Clean up the container after the test is complete
//	t.Cleanup(func() {
//		if err := container.Terminate(ctx); err != nil {
//			t.Fatalf("failed to terminate container: %s", err)
//		}
//	})
//
//	// perform assertions
//
//	host, _ := container.Host(ctx)
//
//	p, _ := container.MappedPort(ctx, "3306/tcp")
//	port := p.Int()
//
//	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?tls=skip-verify",
//		container.Username(), container.Password(), host, port, container.Database())
//
//	db, err := sql.Open("mysql", connectionString)
//	if err != nil {
//		t.Fatal(err)
//	}
//	defer db.Close()
//
//	if err = db.Ping(); err != nil {
//		t.Errorf("error pinging db: %+v\n", err)
//	}
//	stmt, _ := db.Prepare("SELECT @@GLOBAL.innodb_file_format")
//	row := stmt.QueryRow()
//	var innodbFileFormat = ""
//	err = row.Scan(&innodbFileFormat)
//	if err != nil {
//		t.Errorf("error fetching innodb_file_format value")
//	}
//	if innodbFileFormat != "Barracuda" {
//		t.Fatal("The InnoDB file format has been set by the ini file content")
//	}
//}
