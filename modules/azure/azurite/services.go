package azurite

const (
	// blobService is the service name for the Blob service
	blobService Service = "blob"

	// Deprecated: this constant is kept for backward compatibility, but it'll be removed in the next major version.
	// BlobService is the service name for the Blob service
	BlobService Service = blobService

	// queueService is the service name for the Queue service
	queueService Service = "queue"

	// Deprecated: this constant is kept for backward compatibility, but it'll be removed in the next major version.
	// QueueService is the service name for the Queue service
	QueueService Service = queueService

	// tableService is the service name for the Table service
	tableService Service = "table"

	// Deprecated: this constant is kept for backward compatibility, but it'll be removed in the next major version.
	// TableService is the service name for the Table service
	TableService Service = tableService
)

// Deprecated: this type is kept for backward compatibility, but it'll be removed in the next major version.
// Service is the type for the services that Azurite can provide
type Service = service

// service is the type for the services that Azurite can provide
type service string
