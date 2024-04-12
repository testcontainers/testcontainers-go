# TLS certificates

Interacting with services that require TLS certificates is a common issue when working with containers. You can create one or more on-the-fly certificates in order to communicate with your services.

_Testcontainers for Go_ uses a library to generate certificates on-the-fly. This library is called [tlscert](https://github.com/mdelapenya/tlscert).

### Examples

In the following example we are going to start an HTTP server with a self-signed certificate.
It exposes one single handler that will return a simple message when accessed.
The example will also create a client that will connect to the server using the generated certificate,
demonstrating how to use the generated certificate to communicate with a service.

<!--codeinclude-->
[Create a self-signed certificate](../../modules/cockroachdb/certs.go) inside_block:exampleSelfSignedCert
[Sign a self-signed certificate](../../modules/cockroachdb/certs.go) inside_block:exampleSignSelfSignedCert
<!--/codeinclude-->
