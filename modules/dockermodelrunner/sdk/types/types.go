package types

// ModelResponse is the response for a model
type ModelResponse struct {
	// ID is the ID of the model
	ID string `json:"id"`

	// Tags are the tags of the model
	Tags []string `json:"tags"`

	// Config is the config of the model
	Config map[string]string `json:"config"`

	// Created is the creation time of the model
	Created int64 `json:"created"`
}
