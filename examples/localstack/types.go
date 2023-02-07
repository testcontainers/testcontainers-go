package localstack

import (
	"fmt"
	"strings"

	"github.com/imdario/mergo"
	"github.com/testcontainers/testcontainers-go"
)

type LocalStackContainerRequest struct {
	testcontainers.ContainerRequest
	legacyMode      bool
	region          string
	version         string
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

type localStackContainerOption func(req *LocalStackContainerRequest)

// WithDefaultRegion uses the default region for the container, which is "us-east-1"
var WithDefaultRegion = WithRegion(defaultRegion)

// WithCredentials returns a function that can be used to configure the AWS credentials of the container
func WithCredentials(c Credentials) func(req *LocalStackContainerRequest) {
	return func(req *LocalStackContainerRequest) {
		if req.Env == nil {
			req.Env = map[string]string{}
		}

		if c.AccessKeyID != "" {
			req.Env["AWS_ACCESS_KEY_ID"] = c.AccessKeyID
		}
		if c.SecretAccessKey != "" {
			req.Env["AWS_SECRET_ACCESS_KEY"] = c.SecretAccessKey
		}
		if c.Token != "" {
			req.Env["AWS_SESSION_TOKEN"] = c.Token
		}
	}
}

// WithRegion returns a function that can be used to configure the AWS region of the container
func WithRegion(region string) func(req *LocalStackContainerRequest) {
	return func(req *LocalStackContainerRequest) {
		if req.Env == nil {
			req.Env = map[string]string{}
		}

		if region == "" {
			region = defaultRegion
		}

		req.Env["DEFAULT_REGION"] = region
		req.region = region
	}
}

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

// WithDefaultVersion uses the default version for the container, which is "0.11"
var WithDefaultVersion = WithVersion(defaultVersion)

// WithVersion returns a function that can be used to configure the version of the container
func WithVersion(v string) func(req *LocalStackContainerRequest) {
	return func(req *LocalStackContainerRequest) {
		if v == "" {
			v = defaultVersion
		}

		req.version = v
	}
}

type overrideContainerRequestOption func(req testcontainers.ContainerRequest) testcontainers.ContainerRequest

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
