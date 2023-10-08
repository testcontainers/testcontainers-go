# Multi Wait strategy

The Multi wait strategy holds a list of wait strategies. The execution of each strategy is first added, first executed.

Available Options:

- `WithDeadline` - the deadline for when all strategies must complete by, default is none.
- `WithStartupTimeoutDefault` - the startup timeout default to be used for each Strategy if not defined in seconds, default is 60 seconds.

```golang
req := ContainerRequest{
    Image:        "docker.io/mysql:8.0.30",
    ExposedPorts: []string{"3306/tcp", "33060/tcp"},
    Env: map[string]string{
        "MYSQL_ROOT_PASSWORD": "password",
        "MYSQL_DATABASE":      "database",
    },
    wait.ForAll(
          wait.ForLog("port: 3306  MySQL Community Server - GPL"),              // Timeout: 120s (from ForAll.WithStartupTimeoutDefault)
          wait.ForExposedPort().WithStartupTimeout(180*time.Second),            // Timeout: 180s
          wait.ForListeningPort("3306/tcp").WithStartupTimeout(10*time.Second), // Timeout: 10s
    ).WithStartupTimeoutDefault(120*time.Second).                               // Applies default StartupTimeout when not explicitly defined
      WithDeadline(360*time.Second)                                             // Applies deadline for all Wait Strategies
}
```
