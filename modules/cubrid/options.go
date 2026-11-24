package cubrid

import (
	"errors"
	"fmt"
	"strings"

	"github.com/testcontainers/testcontainers-go"
)

// WithDatabase sets the name of the database created on startup. The default is testdb.
func WithDatabase(name string) testcontainers.CustomizeRequestOption {
	return testcontainers.WithEnv(map[string]string{"CUBRID_DB": name})
}

// WithUsername sets the user created on startup. The default is testuser.
// It cannot be dba or public, which already exist in CUBRID.
func WithUsername(user string) testcontainers.CustomizeRequestOption {
	return testcontainers.WithEnv(map[string]string{"CUBRID_USER": user})
}

// WithPassword sets the password of the user created on startup. The default is testpass.
func WithPassword(password string) testcontainers.CustomizeRequestOption {
	return testcontainers.WithEnv(map[string]string{"CUBRID_PASSWORD": password})
}

// WithLocale sets the locale the database is created with, such as ko_KR.utf8.
// Without it the image creates an en_US database, whose charset is ISO-8859-1.
func WithLocale(locale string) testcontainers.CustomizeRequestOption {
	return testcontainers.WithEnv(map[string]string{"CUBRID_LOCALE": locale})
}

func validateSettings(req *testcontainers.GenericContainerRequest) error {
	if req.Env["CUBRID_DB"] == "" {
		return errors.New("database must not be empty")
	}

	user := req.Env["CUBRID_USER"]
	switch strings.ToLower(user) {
	case "":
		return errors.New("username must not be empty")
	case "dba", "public":
		return fmt.Errorf("username %q is reserved by CUBRID", user)
	}

	if locale, ok := req.Env["CUBRID_LOCALE"]; ok && locale == "" {
		return errors.New("locale must not be empty")
	}

	return nil
}
