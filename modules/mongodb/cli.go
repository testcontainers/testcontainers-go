package mongodb

import "fmt"

// mongoCli is cli to interact with MongoDB. If username and password are provided
// it will use credentials to authenticate.
type mongoCli struct {
	mongoshBaseCmd string
	mongoBaseCmd   string
}

func newMongoCli(username string, password string) mongoCli {
	authArgs := ""
	if username != "" && password != "" {
		authArgs = fmt.Sprintf("--username %s --password %s", username, password)
	}

	return mongoCli{
		mongoshBaseCmd: fmt.Sprintf("mongosh %s --quiet", authArgs),
		mongoBaseCmd:   fmt.Sprintf("mongo %s --quiet", authArgs),
	}
}

func (m mongoCli) eval(command string, args ...any) []string {
	command = "\"" + fmt.Sprintf(command, args...) + "\""

	return []string{
		"sh",
		"-c",
		m.mongoshBaseCmd + " --eval " + command + " || " + m.mongoBaseCmd + " --eval " + command,
	}
}
