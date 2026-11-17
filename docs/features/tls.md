# TLS certificates

Interacting with services that require TLS certificates is a common issue when working with containers. You can create one or more on-the-fly certificates in order to communicate with your services.

_Testcontainers for Go_ uses a library to generate certificates on-the-fly. This library is called [tlscert](https://github.com/mdelapenya/tlscert).

### Examples

The following executable example creates a self-signed certificate authority (CA),
uses it to sign a server certificate, and starts a local HTTPS server with that certificate.
The client trusts the generated CA through `RootCAs` and verifies the server's certificate
and hostname before reading the response. The server certificate includes `127.0.0.1`
and `localhost` as subject alternative names to match the local server address.

The example uses Go's `httptest` package, so it can run without Docker. For a containerized
service, configure the service with the generated server certificate and private key,
and use the same client trust configuration. Include the hostname or IP used by the client
in the server certificate's `Host` field.

<!--codeinclude-->
[HTTPS server and client](../../modules/rabbitmq/tls_example_test.go) inside_block:exampleTLSCertificates
<!--/codeinclude-->

The response is `Hello over TLS!`. See the [RabbitMQ SSL example](../modules/rabbitmq.md)
for configuring certificates on a containerized service.

To run this example from the repository root:

```shell
cd modules/rabbitmq
go test -run '^Example_tlsCertificates$' -v
```
