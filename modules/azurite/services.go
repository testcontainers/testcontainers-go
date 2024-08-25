package azurite

const (
	BlobService  Service = "blob"
	QueueService Service = "queue"
	TableService Service = "table"
)

type Service string
