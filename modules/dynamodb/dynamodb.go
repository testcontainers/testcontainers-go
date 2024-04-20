package dynamodb

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"k8s.io/utils/ptr"
)

const (
	DefaultPort        = "8080"
	DefaultExposedPort = DefaultPort + "/tcp"
)

// DynamoDBContainer represents the DynamoDB container type used in the module
type DynamoDBContainer struct {
	testcontainers.Container

	// endpointURL is the fully-qualified URL for the local DynamoDB endpoint
	endpointURL string
}

// RunContainer creates an instance of the DynamoDB container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*DynamoDBContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "amazon/dynamodb-local:2.4.0",
		ExposedPorts: []string{DefaultExposedPort},
		WaitingFor:   wait.NewHostPortStrategy(DefaultExposedPort),
		Cmd: []string{
			"-jar",
			"DynamoDBLocal.jar",
			"-inMemory",
			"-port", DefaultPort,
			"-disableTelemetry",
		},
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		opt.Customize(&genericContainerReq)
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	if err != nil {
		return nil, err
	}

	dynamoContainer := &DynamoDBContainer{
		Container: container,
	}
	return dynamoContainer, dynamoContainer.configureContainer(ctx, newOptions(opts...))
}

// DynamoDBClient creates a new AWS SDK v2 DynamoDB client
func (d *DynamoDBContainer) DynamoDBClient(ctx context.Context) (*dynamodb.Client, error) {
	cnf, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion("us-east-1"),
		config.WithEndpointResolverWithOptions(d.newEndpointResolverWithOptionsFunc()),
		config.WithCredentialsProvider(d.newCredentialsProvider()),
	)
	if err != nil {
		return nil, err
	}
	return dynamodb.NewFromConfig(cnf), nil
}

// configureContainer sets up internal container config & applies user-supplied options.
func (d *DynamoDBContainer) configureContainer(ctx context.Context, config options) error {
	// set the endpointURL field
	if err := d.setEndpointURL(ctx); err != nil {
		return err
	}

	// create tables
	if err := d.createTables(ctx, config.createTables...); err != nil {
		return fmt.Errorf("error creating tables: %w", err)
	}

	return nil
}

// setEndpointURL sets the endpointURL field of the DynamoDB container struct
func (d *DynamoDBContainer) setEndpointURL(ctx context.Context) error {
	mappedPort, err := d.Container.MappedPort(ctx, DefaultExposedPort)
	if err != nil {
		return fmt.Errorf("error getting mapped port: %w", err)
	}

	hostname, err := d.Container.Host(ctx)
	if err != nil {
		return fmt.Errorf("error getting hostname: %w", err)
	}

	d.endpointURL = fmt.Sprintf("http://%s:%s", hostname, mappedPort.Port())
	return nil
}

// createTables is a convenience method to create one or more DynamoDB tables.
func (d *DynamoDBContainer) createTables(ctx context.Context, tables ...dynamodb.CreateTableInput) error {
	cl, err := d.DynamoDBClient(ctx)
	if err != nil {
		return fmt.Errorf("error getting DynamoDB client: %w", err)
	}

	for i, table := range tables {
		if _, err := cl.CreateTable(ctx, &table); err != nil {
			return fmt.Errorf("error creating table %s: %w", ptr.Deref(table.TableName, strconv.Itoa(i)), err)
		}
	}
	return nil
}

// newEndpointResolverWithOptionsFunc generates an AWS SDK v2 EndpointResolverWithOptionsFunc for the DynamoDB endpoint
func (d *DynamoDBContainer) newEndpointResolverWithOptionsFunc() aws.EndpointResolverWithOptionsFunc {
	return func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               d.endpointURL,
			HostnameImmutable: true,
		}, nil
	}
}

// newCredentialsProvider generates an AWS SDK v2 CredentialsProvider for the DynamoDB endpoint
func (d *DynamoDBContainer) newCredentialsProvider() aws.CredentialsProvider {
	return credentials.StaticCredentialsProvider{
		Value: aws.Credentials{
			AccessKeyID:     "FAKE",
			SecretAccessKey: "FAKE",
		},
	}
}
