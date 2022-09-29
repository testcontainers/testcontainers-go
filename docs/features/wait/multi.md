# Multi Wait strategy

The Multi wait strategy will hold a list of wait strategies, in order to wait for all of them. It's possible to set the following conditions:

- the startup timeout to be used in seconds, default is 60 seconds.

```golang
req := ContainerRequest{
    Image:        "docker.io/mysql:latest",
    ExposedPorts: []string{"3306/tcp", "33060/tcp"},
    Env: map[string]string{
        "MYSQL_ROOT_PASSWORD": "password",
        "MYSQL_DATABASE":      "database",
    },
    WaitingFor: wait.ForAll(
        wait.ForLog("port: 3306  MySQL Community Server - GPL"),
        wait.ForListeningPort("3306/tcp"),
    ).WithStartupTimeout(10*time.Second),
}
```
