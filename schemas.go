package core

import (
	"bytes"
	"encoding/json"
	"time"

	"context"
	"errors"
	"io"
	"log"
	"net/http"
)

var (
	schemaVersions = map[string]bool{"0.1": true}
	modelVersions  = map[string]bool{"1.0": true}
)

type requestError struct {
	SchemaVersionError string   `json:"schema_version_error,omitempty"`
	ModelVersionError  string   `json:"model_version_error,omitempty"`
	TimestampError     string   `json:"timestamp_error,omitempty"`
	DataError          string   `json:"data_error,omitempty"`
	UnknownFields      []string `json:"unknown_fields,omitempty"`
}

func (re *requestError) Error() string {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(re)
	if err != nil {
		buf.WriteString("Request doesn't comply with schema")
	}
	return buf.String()
}

// Request represents user request
type Request struct {
	SchemaVersion string                 `json:"schema_version"`
	ModelVersion  string                 `json:"model_version"`
	Timestamp     time.Time              `json:"timestamp"`
	Data          map[string]interface{} `json:"data"`
}

// UnmarshalJSON implements custom json unmarshalling with validation
// TODO: separate validation
func (r *Request) UnmarshalJSON(data []byte) error {
	log.Printf("Request: %s", data)
	tmp := make(map[string]interface{})
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	rerr := &requestError{}
	raise := false
	if v, ok := tmp["schema_version"]; ok {
		r.SchemaVersion = v.(string)
		delete(tmp, "schema_version")
	} else {
		rerr.SchemaVersionError = "No schema version"
		raise = true
	}
	if v, ok := tmp["model_version"]; ok {
		r.ModelVersion = v.(string)
		delete(tmp, "model_version")
	} else {
		rerr.ModelVersionError = "No model version"
		raise = true
	}
	if v, ok := tmp["timestamp"]; ok {
		tt, err := time.Parse(time.RFC3339, v.(string))
		if err != nil {
			rerr.TimestampError = "Wrong timestamp format, use RFC3339"
			raise = true
		}
		r.Timestamp = tt
		delete(tmp, "timestamp")
	} else {
		rerr.TimestampError = "No timestamp"
		raise = true
	}
	if v, ok := tmp["data"]; ok {
		r.Data = v.(map[string]interface{})
		delete(tmp, "data")
	}
	for k := range tmp {
		rerr.UnknownFields = append(rerr.UnknownFields, k)
	}
	if len(rerr.UnknownFields) > 0 {
		raise = true
	}
	if !schemaVersions[r.SchemaVersion] {
		rerr.SchemaVersionError = "Wrong schema version"
		raise = true
	}
	if !modelVersions[r.ModelVersion] {
		rerr.ModelVersionError = "Wrong model version"
		raise = true
	}
	if raise {
		return rerr
	}
	return nil
}

// DataString encodes request data to string
func (r *Request) DataString() string {
	if len(r.Data) == 0 {
		return ""
	}
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(r.Data)
	if err != nil {
		return ""
	}
	return buf.String()
}

// AppError is main app error type
type AppError struct {
	Err        error
	StatusCode int
	Reason     string
	Stacktrace string
}

func (ae *AppError) Error() string {
	if ae.Reason != "" {
		return ae.Reason
	}
	if ae.Err != nil {
		return ae.Err.Error()
	}
	return http.StatusText(ae.StatusCode)
}

// Write writes response to http.ResponseWriter with given context
func (ae *AppError) Write(ctx context.Context, w http.ResponseWriter) {
	resp := &Response{
		SchemaVersion: "0.1",
		ModelVersion:  "1.0",
		Timestamp:     time.Now().UTC(),
		Stacktrace:    ae.Stacktrace,
		Status:        "error",
		Reason:        ae.Error(),
	}
	w.WriteHeader(ae.StatusCode)
	resp.Write(ctx, w)
}

// CreateResponseFromRequest creates response object from user request
func CreateResponseFromRequest(r *Request) *Response {
	return &Response{
		SchemaVersion: r.SchemaVersion,
		ModelVersion:  r.ModelVersion,
	}
}

// Response is main response object
type Response struct {
	SchemaVersion string          `json:"schema_version"`
	ModelVersion  string          `json:"model_version"`
	Timestamp     time.Time       `json:"timestamp"`
	Status        string          `json:"status"`
	ExecutionTime time.Duration   `json:"execution_time,omitempty"`
	Data          json.RawMessage `json:"data,omitempty"`
	Reason        string          `json:"reason,omitempty"`
	Stacktrace    string          `json:"stacktrace,omitempty"`
	err           *AppError
}

// MarshalJSON implements custom json marshalling with logging
func (sr *Response) MarshalJSON() ([]byte, error) {
	type ResponseAlias Response
	if sr.err != nil {
		sr.ExecutionTime = 0
		sr.Data = []byte{}
	}
	resp, err := json.Marshal(&struct{ *ResponseAlias }{(*ResponseAlias)(sr)})
	if err != nil {
		log.Printf("Error encoding response: %s\n%s", err, sr)
	}
	log.Printf("Response: %s", resp)
	return resp, err
}

// Write writes response to http.ResponseWriter with given context
func (sr *Response) Write(ctx context.Context, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	sr.Timestamp = time.Now().UTC()
	err := json.NewEncoder(w).Encode(sr)
	if err != nil {
		if sr.err == nil {
			sr.err = &AppError{
				StatusCode: http.StatusInternalServerError,
				Reason:     "Script does not return valid json string.",
			}
		}
		sr.err.Write(ctx, w)
	}
}

// BuildRequest builds and validates user requests
func BuildRequest(r io.Reader) (*Request, error) {
	var request Request
	err := json.NewDecoder(r).Decode(&request)
	if err != nil {
		if err == io.EOF {
			return nil, errors.New("received empty request")
		}
		return nil, errors.New("Request doesn't comply with schema")
	}
	return &request, nil
}
