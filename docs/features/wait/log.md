# Log Wait strategy

The Log wait strategy will check if a string occurs in the container logs for a desired number of times, and allows to set the following conditions:

- the string to be waited for in the container log.
- the number of occurrences of the string to wait for, default is `1`.
- look for the string using a regular expression, default is `false`.
- the startup timeout to be used in seconds, default is 60 seconds.
- the poll interval to be used in milliseconds, default is 100 milliseconds.

```golang
req := ContainerRequest{
    Image:        "mysql:8.0.36",
    ExposedPorts: []string{"3306/tcp", "33060/tcp"},
    Env: map[string]string{
        "MYSQL_ROOT_PASSWORD": "password",
        "MYSQL_DATABASE":      "database",
    },
    WaitingFor: wait.ForLog("port: 3306  MySQL Community Server - GPL"),
}
```

Using a regular expression:

```golang
req := ContainerRequest{
    Image:        "mysql:8.0.36",
    ExposedPorts: []string{"3306/tcp", "33060/tcp"},
    Env: map[string]string{
        "MYSQL_ROOT_PASSWORD": "password",
        "MYSQL_DATABASE":      "database",
    },
    WaitingFor: wait.ForLog(`.*MySQL Community Server`).AsRegexp(),
}
```
