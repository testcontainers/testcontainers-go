# TLS certificates

_Testcontainers for Go_ provides a way to interact with services that require TLS certificates.
You can create one or more on-the-fly certificates in order to communicate with your services.

To create a certificate, you can use the `tctls.Certificate` struct, where the `tctls` namespace is imported from the `github.com/testcontainers/testcontainers-go/tls` package, to avoid conflicts with the `tls` package from the Go standard library.

The `tctls.Certificate` struct has the following fields:

<!--codeinclude-->
[Certificate struct](../../tls/generate.go) inside_block:testcontainersTLSCertificate
<!--/codeinclude-->

You can generate a certificate by calling the `tctls.GenerateCert` function. This function receives a variadic argument of functional options that allow you to customize the certificate:

- `tctls.WithSubjectCommonName`: sets the subject's common name of the certificate.
- `tctls.WithHost`: sets the hostnames that the certificate will be valid for. In the case the passed string contains comma-separated values,
it will be split into multiple hostnames and IPs. Each hostname and IP will be trimmed of whitespace, and if the value is an IP,
it will be added to the IPAddresses field of the certificate, after the ones passed with the WithIPAddresses option.
Otherwise, it will be added to the DNSNames field.
- `tctls.WithValidFor`: sets the duration that the certificate will be valid for. By default, the certificate is valid for 365 days.
- `tctls.AsCA`: sets the certificate as a Certificate Authority (CA). This option is disabled by default.
When passed, the KeyUsage field of the certificate will append the x509.KeyUsageCertSign usage.
- `tctls.WithParent`: sets the parent certificate of the certificate. This option is disabled by default.
When passed, the parent certificate will be used to sign the generated certificate,
and the issuer of the certificate will be set to the common name of the parent certificate.
- `tctls.AsPem`: sets the certificate to be returned as PEM bytes. It will include the private key in the `KeyBytes` field of the Certificate struct.
- `tctls.WithIPAddresses`: sets the IP addresses that the certificate will be valid for. The IPs passed with this option will be added
first to the IPAddresses field of the certificate: those coming from the `WithHost` option will be added after them.
- `tctls.WithSaveToFile`: sets the parent directory where the certificate and its private key will be saved. Both the certificate and its private key will be saved in separate files, using an random UUID as part of the filename. E.g., `cert-<UUID>.pem` and `key-<UUID>.pem`.

!!! note
    If the `WithSaveToFile` option is passed, it will automatically set the `AsPem` option, as we need to the private key bytes too.

### Examples

In the following example we are going to start an HTTP server with a self-signed certificate.
It exposes one single handler that will return a simple message when accessed.
The example will also create a client that will connect to the server using the generated certificate,
demonstrating how to use the generated certificate to communicate with a service.

<!--codeinclude-->
[Use a certificate](../../tls/examples_test.go) inside_block:ExampleGenerateCert
<!--/codeinclude-->
