package localstack

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/go-connections/nat"
	"github.com/imdario/mergo"
	"github.com/testcontainers/testcontainers-go"
)

// LocalStackContainer represents the LocalStack container type used in the module
type LocalStackContainer struct {
	testcontainers.Container
	EnabledServices map[string]Service
}

// ServicePort returns the port of the given service
func (l *LocalStackContainer) ServicePort(ctx context.Context, service EnabledService) (nat.Port, error) {
	if _, ok := l.EnabledServices[service.Name()]; !ok {
		return "", fmt.Errorf("service %s is not enabled", service.Name())
	}

	internalPort := l.EnabledServices[service.Name()].servicePort()

	p, err := nat.NewPort("tcp", fmt.Sprintf("%d", internalPort))
	if err != nil {
		return "", err
	}

	return l.MappedPort(ctx, p)
}

type LocalStackContainerRequest struct {
	testcontainers.ContainerRequest
	legacyMode      bool
	enabledServices []Service
}

// EnabledService is an interface for services that can be enabled
type EnabledService interface {
	Name() string
	Named(string) string
	Port() int
}

type Service struct {
	name       string
	legacyMode bool
	port       int
}

// Name returns the name of the service
func (s Service) Name() string {
	return s.name
}

// Named returns the name of the service with the given name
func (s Service) Named(name string) string {
	return name
}

// Port returns the port of the service
func (s Service) Port() int {
	return s.port
}

func (s Service) servicePort() int {
	if s.legacyMode {
		return s.Port()
	}

	return defaultPort
}

// List of AWS services, ordered alphabetically

// APIGateway is the API Gateway service
var APIGateway = Service{
	name: "apigateway",
	port: 4567,
}

// CloudFormation is the CloudFormation service
var CloudFormation = Service{
	name: "cloudformation",
	port: 4581,
}

// CloudWatch is the CloudWatch service
var CloudWatch = Service{
	name: "cloudwatch",
	port: 4582,
}

// CloudWatchLogs is the CloudWatchLogs service
var CloudWatchLogs = Service{
	name: "cloudwatchlogs",
	port: 4586,
}

// DynamoDB is the DynamoDB service
var DynamoDB = Service{
	name: "dynamodb",
	port: 4569,
}

// DynamoDBStreams is the DynamoDB Streams service
var DynamoDBStreams = Service{
	name: "dynamodbstreams",
	port: 4570,
}

// EC2 is the EC2 service
var EC2 = Service{
	name: "ec2",
	port: 4597,
}

// Firehose is the Firehose service
var Firehose = Service{
	name: "firehose",
	port: 4573,
}

// IAM is the IAM service
var IAM = Service{
	name: "iam",
	port: 4593,
}

// KMS is the KMS service
var KMS = Service{
	name: "kms",
	port: 4599,
}

// Kinesis is the Kinesis service
var Kinesis = Service{
	name: "kinesis",
	port: 4568,
}

// Lambda is the Lambda service
var Lambda = Service{
	name: "lambda",
	port: 4574,
}

// Redshift is the Redshift service
var Redshift = Service{
	name: "redshift",
	port: 4577,
}

// Route53 is the Route53 service
var Route53 = Service{
	name: "route53",
	port: 4580,
}

// S3 is the S3 service
var S3 = Service{
	name: "s3",
	port: 4572,
}

// SES is the SES service
var SES = Service{
	name: "ses",
	port: 4579,
}

// SNS is the SNS service
var SNS = Service{
	name: "sns",
	port: 4575,
}

// SQS is the SQS service
var SQS = Service{
	name: "sqs",
	port: 4576,
}

// SSM is the SSM service
var SSM = Service{
	name: "ssm",
	port: 4583,
}

// STS is the STS service
var STS = Service{
	name: "sts",
	port: 4592,
}

// SecretsManager is the SecretsManager service
var SecretsManager = Service{
	name: "secretsmanager",
	port: 4584,
}

// StepFunctions is the StepFunctions service
var StepFunctions = Service{
	name: "stepfunctions",
	port: 4585,
}

type LocalStackContainerOption func(req *LocalStackContainerRequest)

// WithLegacyMode uses the legacy mode for the container, which exposes each service on a different port
var WithLegacyMode = func(req *LocalStackContainerRequest) {
	req.legacyMode = true
}

// WithServices returns a function that can be used to configure the container
func WithServices(services ...Service) func(req *LocalStackContainerRequest) {
	return func(req *LocalStackContainerRequest) {
		serviceNames := []string{}
		servicesPorts := map[int]bool{} // map of unique exposed ports

		for _, service := range services {
			// automatically set legacy mode for services that require it,
			// as it will affect how the service is exposed
			service.legacyMode = req.legacyMode

			servicePort := service.servicePort()

			servicesPorts[servicePort] = true
			serviceNames = append(serviceNames, service.Name())

			// add services to the list of enabled services
			req.enabledServices = append(req.enabledServices, service)
		}

		if req.Env == nil {
			req.Env = map[string]string{}
		}

		if len(serviceNames) > 0 {
			req.Env["SERVICES"] = strings.Join(serviceNames, ",")
		}

		exposedPorts := []string{}
		for port := range servicesPorts {
			exposedPorts = append(exposedPorts, fmt.Sprintf("%d/tcp", port))
		}

		req.ExposedPorts = exposedPorts
	}
}

type OverrideContainerRequestOption func(req testcontainers.ContainerRequest) testcontainers.ContainerRequest

// NoopOverrideContainerRequest returns a function that can be used to be merged with the container request
var NoopOverrideContainerRequest = func(req testcontainers.ContainerRequest) testcontainers.ContainerRequest {
	return req
}

// OverrideContainerRequest returns a function that can be used to be merged with the container request
func OverrideContainerRequest(r testcontainers.ContainerRequest) func(req testcontainers.ContainerRequest) testcontainers.ContainerRequest {
	return func(req testcontainers.ContainerRequest) testcontainers.ContainerRequest {
		if err := mergo.Merge(&req, r, mergo.WithOverride); err != nil {
			fmt.Printf("error merging container request %v. Keeping the default one: %v", err, req)
			return req
		}

		return req
	}
}
