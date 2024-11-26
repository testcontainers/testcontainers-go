# TLS Strategy

TLS Strategy waits for one or more files to exist in the container and uses them
and other details to construct a `tls.Config` which can be used to create secure
connections.

It supports:

- x509 PEM Certificate loaded from a certificate / key file pair.
- Root Certificate Authorities aka RootCAs loaded from PEM encoded files.
- Server name.
- Startup timeout to be used in seconds, default is 60 seconds.
- Poll interval to be used in milliseconds, default is 100 milliseconds.

## Waiting for certificate pair

The following snippets show how to configure a request to wait for certificate
pair to exist once started and then read the
[tls.Config](https://pkg.go.dev/crypto/tls#Config), alongside how to copy a test
certificate pair into a container image using a `Dockerfile`.

It should be noted that copying certificate pairs into an images is only an
example which might be useful for testing with testcontainers-go and should not
be done with production images as that could expose your certificates if your
images become public.

<!--codeinclude-->
[Wait for certificate](../../../wait/tls_test.go) inside_block:waitForTLSCert
[Read TLS Config](../../../wait/tls_test.go) inside_block:waitTLSConfig
[Dockerfile with certificate](../../../wait/testdata/http/Dockerfile)
<!--/codeinclude-->
