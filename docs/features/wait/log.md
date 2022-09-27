# Log Wait strategy

The Log wait strategy will check if a string is present in the container logs with a desired number of ocurrences, being able to set the following conditions:

- the string to be searched in the container log.
- the number of ocurrences of the searched string, default is `1`.
- the startup timeout to be used, default is 60 seconds.
- the poll interval to be used, default is 100 milliseconds.

```golang
req := ContainerRequest{
    Image:        "docker.io/mysql:latest",
    ExposedPorts: []string{"3306/tcp", "33060/tcp"},
    Env: map[string]string{
        "MYSQL_ROOT_PASSWORD": "password",
        "MYSQL_DATABASE":      "database",
    },
    WaitingFor: wait.ForLog("port: 3306  MySQL Community Server - GPL"),
}
```
