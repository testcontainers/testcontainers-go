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
data_transforms_per_function_memory_limit: 16777216
data_transforms_per_core_memory_reservation: 33554432
{{- end }}

{{- if .AutoCreateTopics }}
auto_create_topics_enabled: true
{{- end }}
