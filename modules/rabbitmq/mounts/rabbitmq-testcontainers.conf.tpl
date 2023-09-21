# Injected by testcontainers
default_user = {{ .AdminUsername }}
default_pass = {{ .AdminPassword }}

{{- if .SSLSettings }}
listeners.tcp = none
listeners.ssl.default = 5671
ssl_options.cacertfile = /etc/rabbitmq/ca_cert.pem
ssl_options.certfile = /etc/rabbitmq/rabbitmq_cert.pem
ssl_options.keyfile = /etc/rabbitmq/rabbitmq_key.pem
ssl_options.depth = {{ .SSLSettings.VerificationDepth }}
ssl_options.verify = {{ .SSLSettings.VerificationMode }}
ssl_options.fail_if_no_peer_cert = {{ .SSLSettings.FailIfNoCert }}
{{- end }}
