# Injected by testcontainers
# This file contains cluster properties which will only be considered when
# starting the cluster for the first time. Afterwards, you can configure cluster
# properties via the Redpanda Admin API.
superusers:
{{- if not .Superusers }}
  []
{{- else }}
{{- range .Superusers }}
  - {{.}}
{{- end }}
{{- end }}

{{- if .KafkaAPIEnableAuthorization }}
kafka_enable_authorization: true
{{- end }}

{{- if .EnableWasmTransform }}
data_transforms_enabled: true
{{- end }}

{{- if .AutoCreateTopics }}
auto_create_topics_enabled: true
{{- end }}

{{- range $key, $value := .ExtraBootstrapConfig }}
{{ $key }}: {{ $value }}
{{- end }}

