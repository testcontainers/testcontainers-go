# HTTP(S) Wait strategy

You can choose to wait for an HTTP(S) endpoint that runs in the container, being able to set the following conditions:

- the port to be used
- the path to be used
- the HTTP method to be used
- the HTTP request body to be sent
- the HTTP status code as a function to resolve a matcher
- the HTTP response as a function to resolve a matcher
- the TLS config to be used for HTTPS
- the PollInterval to be used, default is 100 milliseconds

Variations on the HTTP wait strategy are supported, including:

## Waiting for an HTTP endpoint matching an HTTP method

```golang
req := ContainerRequest{
		Image:        "influxdb:1.8.10-alpine",
		ExposedPorts: []string{"8086/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForHTTP("/ping").WithMethod(http.MethodPost).WithBody(bytes.NewReader([]byte("ping"))),
		),
	}
```

## Waiting for an HTTP endpoint matching an HTTP status code

```golang
req := ContainerRequest{
		Image:        "influxdb:1.8.10-alpine",
		ExposedPorts: []string{"8086/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForHTTP("/ping").WithPort("8086/tcp").WithStatusCodeMatcher(
				func(status int) bool {
					return status == http.StatusNoContent
				},
			),
		),
	}
```

## Waiting for an HTTPS endpoint including an HTTPS status code and a response matcher

```golang
req := testcontainers.ContainerRequest{
    FromDockerfile: testcontainers.FromDockerfile{
        Context: workdir + "/testdata",
    },
    ExposedPorts: []string{"80/tcp"},
    WaitingFor: wait.NewHTTPStrategy("/ping").
        WithStartupTimeout(time.Second * 10).WithPort("80/tcp").
        WithResponseMatcher(func(body io.Reader) bool {
            data, _ := ioutil.ReadAll(body)
            return bytes.Equal(data, []byte("pong"))
        }).
        WithStatusCodeMatcher(func(status int) bool {
            i++ // always fail the first try in order to force the polling loop to be re-run
            return i > 1 && status == 200
        }).
        WithMethod(http.MethodPost).WithBody(bytes.NewReader([]byte("ping"))),
}
```
