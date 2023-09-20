# Injected by testcontainers
default_user = {{ .AdminUsername }}
default_pass = {{ .AdminPassword }}

{{- if .SSLSettings }}
ssl_listeners = 5671
ssl_options.cacertfile = {{ .SSLSettings.CACertFile }}
ssl_options.certfile = {{ .SSLSettings.CertFile }}
ssl_options.keyfile = {{ .SSLSettings.KeyFile }}
ssl_options.depth = {{ .SSLSettings.VerificationDepth }}
ssl_options.verify = {{ .SSLSettings.VerificationMode }}
ssl_options.fail_if_no_peer_cert = {{ .SSLSettings.FailIfNoCert }}
{{- end }}
