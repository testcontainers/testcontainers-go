module e2e

go 1.13

require (
	github.com/aws/aws-sdk-go v1.44.99 // indirect
	github.com/docker/go-connections v0.4.0
	github.com/hashicorp/consul/sdk v0.11.0 // indirect
	github.com/lib/pq v1.10.7
	github.com/miekg/dns v1.1.50 // indirect
	github.com/testcontainers/testcontainers-go v0.13.0
	gotest.tools/gotestsum v1.8.2
)

replace github.com/testcontainers/testcontainers-go => ../
