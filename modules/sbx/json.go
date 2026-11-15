package sbx

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

type dockerJSONField struct {
	name      string
	canonical string
	value     json.RawMessage
}

// dockerJSONObject keeps object members in order so unknown duplicate members
// retain their existing JSON semantics. Only fields that the proxy recognizes
// are resolved case-insensitively, checked for ambiguity, and canonicalized.
type dockerJSONObject struct {
	description string
	recognized  []string
	fields      []dockerJSONField
}

type duplicateDockerJSONMemberError struct {
	description string
	canonical   string
	first       string
	second      string
}

func (e *duplicateDockerJSONMemberError) Error() string {
	return fmt.Sprintf("ambiguous %s: JSON members %q and %q both match Docker field %q", e.description, e.first, e.second, e.canonical)
}

func isDuplicateDockerJSONMemberError(err error) bool {
	var duplicate *duplicateDockerJSONMemberError
	return errors.As(err, &duplicate)
}

func newDockerJSONObject(description string, recognized ...string) *dockerJSONObject {
	return &dockerJSONObject{description: description, recognized: append([]string(nil), recognized...)}
}

func decodeDockerJSONObject(raw []byte, description string, recognized ...string) (*dockerJSONObject, error) {
	object := newDockerJSONObject(description, recognized...)
	decoder := json.NewDecoder(bytes.NewReader(raw))
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	delim, ok := token.(json.Delim)
	if !ok || delim != '{' {
		return nil, fmt.Errorf("%s must be a JSON object", description)
	}
	seen := make(map[string]string, len(recognized))
	for decoder.More() {
		nameToken, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		name, ok := nameToken.(string)
		if !ok {
			return nil, fmt.Errorf("%s has a non-string member name", description)
		}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return nil, err
		}
		canonical := canonicalDockerJSONField(name, recognized)
		if canonical != "" {
			if first, exists := seen[canonical]; exists {
				return nil, &duplicateDockerJSONMemberError{
					description: description, canonical: canonical, first: first, second: name,
				}
			}
			seen[canonical] = name
		}
		object.fields = append(object.fields, dockerJSONField{name: name, canonical: canonical, value: append(json.RawMessage(nil), value...)})
	}
	if _, err := decoder.Token(); err != nil {
		return nil, err
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("%s has trailing JSON content", description)
		}
		return nil, err
	}
	return object, nil
}

func canonicalDockerJSONField(name string, recognized []string) string {
	for _, canonical := range recognized {
		if strings.EqualFold(name, canonical) {
			return canonical
		}
	}
	return ""
}

func (o *dockerJSONObject) canonicalize() {
	for i := range o.fields {
		if o.fields[i].canonical != "" {
			o.fields[i].name = o.fields[i].canonical
		}
	}
}

func (o *dockerJSONObject) get(canonical string) (json.RawMessage, bool) {
	for _, field := range o.fields {
		if field.canonical == canonical || (field.canonical == "" && strings.EqualFold(field.name, canonical)) {
			return field.value, true
		}
	}
	return nil, false
}

func (o *dockerJSONObject) set(canonical string, value json.RawMessage) {
	for i := range o.fields {
		if o.fields[i].canonical == canonical || strings.EqualFold(o.fields[i].name, canonical) {
			o.fields[i].name = canonical
			o.fields[i].canonical = canonical
			o.fields[i].value = append(json.RawMessage(nil), value...)
			return
		}
	}
	o.fields = append(o.fields, dockerJSONField{name: canonical, canonical: canonical, value: append(json.RawMessage(nil), value...)})
}

func (o *dockerJSONObject) values() map[string]json.RawMessage {
	values := make(map[string]json.RawMessage, len(o.fields))
	for _, field := range o.fields {
		values[field.name] = field.value
	}
	return values
}

func (o *dockerJSONObject) MarshalJSON() ([]byte, error) {
	var encoded bytes.Buffer
	encoded.WriteByte('{')
	for i, field := range o.fields {
		if i > 0 {
			encoded.WriteByte(',')
		}
		name, err := json.Marshal(field.name)
		if err != nil {
			return nil, err
		}
		encoded.Write(name)
		encoded.WriteByte(':')
		encoded.Write(field.value)
	}
	encoded.WriteByte('}')
	return encoded.Bytes(), nil
}
