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

## Waiting for certificate pair to exist and construct a tls.Config

<!--codeinclude-->
[Waiting for certificate pair to exist and construct a tls.Config](../../../wait/tls_test.go) inside_block:waitForTLSCert
<!--/codeinclude-->
