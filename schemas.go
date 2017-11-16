package main

import (
	"bytes"
	"encoding/json"
	"time"

	"context"
	"errors"
	"io"
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
	SchemaVersion string          `json:"schema_version"`
	ModelVersion  string          `json:"model_version"`
	Timestamp     time.Time       `json:"timestamp"`
	Data          json.RawMessage `json:"data"`
}

// DataString encodes request data to string
func (r *Request) DataString() string {
	return string(r.Data)
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
		logger.Printf("Error encoding response: %s\n%s", err, sr)
	}
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
		return nil, err
	}
	return &request, nil
}
