# HTTP(S) Wait strategy

The HTTP wait strategy will check the result of an HTTP(S) request that is available in the container, being able to set the following conditions:

- the port to be used.
- the path to be used.
- the HTTP method to be used.
- the HTTP request body to be sent.
- the HTTP status code as a function to resolve a matcher.
- the HTTP response as a function to resolve a matcher.
- the TLS config to be used for HTTPS.
- the startup timeout to be used, default is 60 seconds.
- the poll interval to be used, default is 100 milliseconds.

Variations on the HTTP wait strategy are supported, including:

## Match an HTTP method

```golang
req := ContainerRequest{
		Image:        "docker.io/nginx:alpine",
		ExposedPorts: []string{"8086/tcp"},
		WaitingFor: wait.ForHTTP("/ping").WithMethod(http.MethodPost).WithBody(bytes.NewReader([]byte("ping"))),
	}
```

## Match an HTTP status code

```golang
req := ContainerRequest{
		Image:        "docker.io/nginx:alpine",
		ExposedPorts: []string{"8086/tcp"},
		WaitingFor: wait.ForHTTP("/ping").WithPort("8086/tcp").WithStatusCodeMatcher(
            func(status int) bool {
                return status == http.StatusNoContent
            },
        ),
	}
```

## Match an HTTPS status code and a response matcher

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
