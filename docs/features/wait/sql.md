# SQL Wait strategy

The SQL wait strategy will check the result of a SQL query executed in a container representing a SQL database, and allows to set the following conditions:

- the SQL query to be used, default is `SELECT 1`.
- the port to be used.
- the database driver to be used, as a string.
- the URL of the database to be used, as a function returning the URL string.
- the startup timeout to be used in seconds, default is 60 seconds.
- the poll interval to be used in milliseconds, default is 100 milliseconds.

```golang
req := ContainerRequest{
    Image:        "postgres:14.1-alpine",
    ExposedPorts: []string{port},
    Cmd:          []string{"postgres", "-c", "fsync=off"},
    Env:          env,
    WaitingFor: wait.ForSQL(nat.Port(port), "postgres", dbURL).
        WithStartupTimeout(time.Second * 5).
        WithQuery("SELECT 10"),
}
```

Note: You'll also need to import the appropriate [database driver](https://github.com/golang/go/wiki/SQLDrivers) in your test code such that Testcontainers can pick it up when connecting to the database.
