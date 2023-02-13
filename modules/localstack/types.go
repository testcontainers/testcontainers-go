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

// LocalStackContainerRequest represents the LocalStack container request type used in the module
// to configure the container
type LocalStackContainerRequest struct {
	testcontainers.ContainerRequest
	enabledServices []Service
}

// EnabledService is an interface for services that can be enabled
type EnabledService interface {
	Name() string
	Named(string) string
}

// Service represents a LocalStack service, such as S3, SQS, etc.
type Service struct {
	name string
}

// Name returns the name of the service
func (s Service) Name() string {
	return s.name
}

// Named returns the name of the service with the given name
func (s Service) Named(name string) string {
	return name
}

// servicePort returns the default port
func (s Service) servicePort() int {
	return defaultPort
}

// List of AWS services, ordered alphabetically

// APIGateway is the API Gateway service
var APIGateway = Service{
	name: "apigateway",
}

// CloudFormation is the CloudFormation service
var CloudFormation = Service{
	name: "cloudformation",
}

// CloudWatch is the CloudWatch service
var CloudWatch = Service{
	name: "cloudwatch",
}

// CloudWatchLogs is the CloudWatchLogs service
var CloudWatchLogs = Service{
	name: "cloudwatchlogs",
}

// DynamoDB is the DynamoDB service
var DynamoDB = Service{
	name: "dynamodb",
}

// DynamoDBStreams is the DynamoDB Streams service
var DynamoDBStreams = Service{
	name: "dynamodbstreams",
}

// EC2 is the EC2 service
var EC2 = Service{
	name: "ec2",
}

// Firehose is the Firehose service
var Firehose = Service{
	name: "firehose",
}

// IAM is the IAM service
var IAM = Service{
	name: "iam",
}

// KMS is the KMS service
var KMS = Service{
	name: "kms",
}

// Kinesis is the Kinesis service
var Kinesis = Service{
	name: "kinesis",
}

// Lambda is the Lambda service
var Lambda = Service{
	name: "lambda",
}

// Redshift is the Redshift service
var Redshift = Service{
	name: "redshift",
}

// Route53 is the Route53 service
var Route53 = Service{
	name: "route53",
}

// S3 is the S3 service
var S3 = Service{
	name: "s3",
}

// SES is the SES service
var SES = Service{
	name: "ses",
}

// SNS is the SNS service
var SNS = Service{
	name: "sns",
}

// SQS is the SQS service
var SQS = Service{
	name: "sqs",
}

// SSM is the SSM service
var SSM = Service{
	name: "ssm",
}

// STS is the STS service
var STS = Service{
	name: "sts",
}

// SecretsManager is the SecretsManager service
var SecretsManager = Service{
	name: "secretsmanager",
}

// StepFunctions is the StepFunctions service
var StepFunctions = Service{
	name: "stepfunctions",
}

// LocalStackContainerOption is a type that can be used to configure the LocalStack container,
// modifying the LocalStackContainerRequest struct, and the container request that it wraps
type LocalStackContainerOption func(req *LocalStackContainerRequest)

// WithServices returns a function that can be used to configure the container
func WithServices(services ...Service) func(req *LocalStackContainerRequest) {
	return func(req *LocalStackContainerRequest) {
		serviceNames := []string{}
		servicesPorts := map[int]bool{} // map of unique exposed ports

		for _, service := range services {
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

// OverrideContainerRequestOption is a type that can be used to configure the Testcontainers container request.
// The passed request will be merged with the default one.
type OverrideContainerRequestOption func(req testcontainers.ContainerRequest) testcontainers.ContainerRequest

// NoopOverrideContainerRequest returns a helper function that does not override the container request
var NoopOverrideContainerRequest = func(req testcontainers.ContainerRequest) testcontainers.ContainerRequest {
	return req
}

// OverrideContainerRequest returns a function that can be used to merge the passed container request with one that is created by the LocalStack container
func OverrideContainerRequest(r testcontainers.ContainerRequest) func(req testcontainers.ContainerRequest) testcontainers.ContainerRequest {
	return func(req testcontainers.ContainerRequest) testcontainers.ContainerRequest {
		if err := mergo.Merge(&req, r, mergo.WithOverride); err != nil {
			fmt.Printf("error merging container request %v. Keeping the default one: %v", err, req)
			return req
		}

		return req
	}
}
