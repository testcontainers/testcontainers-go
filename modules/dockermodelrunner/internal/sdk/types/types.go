package types

import (
	"encoding/json"
	"time"
)

// ModelResponse is the response for a model
type ModelResponse struct {
	// ID is the ID of the model
	ID string `json:"id"`

	// Tags are the tags of the model
	Tags []string `json:"tags"`

	// Config is the config of the model
	Config map[string]string `json:"config"`

	// Created is the creation time of the model
	Created JSONTime `json:"created"`
}

// JSONTime handles Unix timestamp conversion to/from time.Time
type JSONTime time.Time

// UnmarshalJSON implements json.Unmarshaler
func (j *JSONTime) UnmarshalJSON(b []byte) error {
	var timestamp int64
	if err := json.Unmarshal(b, &timestamp); err != nil {
		return err
	}
	*j = JSONTime(time.Unix(timestamp, 0))
	return nil
}

// MarshalJSON implements json.Marshaler
func (j JSONTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(j).Unix())
}

// Time returns the time.Time representation
func (j JSONTime) Time() time.Time {
	return time.Time(j)
}
