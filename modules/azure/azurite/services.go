package azurite

const (
	// BlobService is the service name for the Blob service
	BlobService Service = "blob"

	// QueueService is the service name for the Queue service
	QueueService Service = "queue"

	// TableService is the service name for the Table service
	TableService Service = "table"
)

// Service is the type for the services that Azurite can provide
type Service string
