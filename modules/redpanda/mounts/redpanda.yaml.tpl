# Injected by testcontainers
redpanda:
  admin:
    address: 0.0.0.0
    port: 9644

  kafka_api:
    - address: 0.0.0.0
      name: external
      port: 9092
      authentication_method: {{ .KafkaAPI.AuthenticationMethod }}

    # This listener is required for the schema registry client. The schema
    # registry client connects via an advertised listener like a normal Kafka
    # client would do. It can't use the other listener because the mapped
    # port is not accessible from within the Redpanda container.
    - address: 0.0.0.0
      name: internal
      port: 9093
      authentication_method: {{ if .KafkaAPI.EnableAuthorization }}sasl{{ else }}none{{ end }}

  {{ range .KafkaAPI.Listeners }}
    - address: 0.0.0.0
      name: {{ .Address }}
      port: {{ .Port }}
      authentication_method: {{ .AuthenticationMethod }}
  {{ end }}

  advertised_kafka_api:
    - address: {{ .KafkaAPI.AdvertisedHost }}
      name: external
      port: {{ .KafkaAPI.AdvertisedPort }}
    - address: 127.0.0.1
      name: internal
      port: 9093
 {{ range .KafkaAPI.Listeners }} 
    - address: {{ .Address }}
      name: {{ .Address }}
      port: {{ .Port }}
  {{ end }} 
  

{{ if .EnableTLS }}
  admin_api_tls:
    - enabled: true
      cert_file: /etc/redpanda/cert.pem
      key_file: /etc/redpanda/key.pem
  kafka_api_tls:
    - name: external
      enabled: true
      cert_file: /etc/redpanda/cert.pem
      key_file: /etc/redpanda/key.pem
{{ end }}

schema_registry:
  schema_registry_api:
  - address: "0.0.0.0"
    name: main
    port: 8081
    authentication_method: {{ .SchemaRegistry.AuthenticationMethod }}

{{ if .EnableTLS }}
  schema_registry_api_tls:
  - name: main
    enabled: true
    cert_file: /etc/redpanda/cert.pem
    key_file: /etc/redpanda/key.pem
{{ end }}

schema_registry_client:
  brokers:
    - address: localhost
      port: 9093

auto_create_topics_enabled: {{ .AutoCreateTopics }}
