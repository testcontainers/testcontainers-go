# Using Docker Compose

Similar to generic containers support, it's also possible to run a bespoke set
of services specified in a docker-compose.yml file.

This is intended to be useful on projects where Docker Compose is already used
in dev or other environments to define services that an application may be
dependent upon.

You can override Testcontainers-go's default behaviour and make it use a
docker-compose binary installed on the local machine. This will generally yield
an experience that is closer to running docker-compose locally, with the caveat
that Docker Compose needs to be present on dev and CI machines.

## Examples

```go
composeFilePaths := []string {"testresources/docker-compose.yml"}
identifier := strings.ToLower(uuid.New().String())

compose := tc.NewLocalDockerCompose(composeFilePaths, identifier)
execError := compose.
	WithCommand([]string{"up", "-d"}).
	WithEnv(map[string]string {
		"key1": "value1",
		"key2": "value2",
	}).
	Invoke()
err := execError.Error
if err != nil {
	return fmt.Errorf("Could not run compose file: %v - %v", composeFilePaths, err)
}
return nil
```

Note that the environment variables in the `env` map will be applied, if
possible, to the existing variables declared in the Docker Compose file.

In the following example, we demonstrate how to stop a Docker Compose created project using the
convenient `Down` method.

```go
composeFilePaths := []string{"testresources/docker-compose.yml"}

compose := tc.NewLocalDockerCompose(composeFilePaths, identifierFromExistingRunningCompose)
execError := compose.Down()
err := execError.Error
if err != nil {
	return fmt.Errorf("Could not run compose file: %v - %v", composeFilePaths, err)
}
return nil
```

