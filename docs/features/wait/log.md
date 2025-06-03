# Log Wait strategy

The Log wait strategy will check if a string occurs in the container logs for a desired number of times, and allows to set the following conditions:

- the string to be waited for in the container log.
- the number of occurrences of the string to wait for, default is `1` (ignored for Submatch).
- look for the string using a regular expression, default is `false`.
- the startup timeout to be used in seconds, default is 60 seconds.
- the poll interval to be used in milliseconds, default is 100 milliseconds.
- the regular expression submatch callback, default nil (occurrences is ignored).

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

Using regular expression with submatch:

```golang
var host, port string
req := ContainerRequest{
    Image:        "ollama/ollama:0.5.7",
    ExposedPorts: []string{"11434/tcp"},
    WaitingFor: wait.ForLog(`Listening on (.*:\d+) \(version\s(.*)\)`).Submatch(func(pattern string, submatches [][][]byte) error {
        var err error
        for _, matches := range submatches {
            if len(matches) != 3 {
                err = fmt.Errorf("`%s` matched %d times, expected %d", pattern, len(matches), 3)
                continue
            }
            host, port, err = net.SplitHostPort(string(matches[1]))
            if err != nil {
                return wait.NewPermanentError(fmt.Errorf("split host port: %w", err))
            }

            // Host and port successfully extracted from log.
            return nil
        }

        if err != nil {
            // Return the last error encountered.
            return err
        }

        return fmt.Errorf("address and version not found: `%s` no matches", pattern)
    }),
}
```

If the return from a Submatch callback function is a `wait.PermanentError` the
wait will stop and the error will be returned. Use `wait.NewPermanentError(err error)`
to achieve this.
