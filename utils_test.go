package main

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"
)

func TestValidateRequest_Success(t *testing.T) {
	r := Request{
		SchemaVersion: "0.1",
		ModelVersion:  "1.0",
		Timestamp:     time.Now().UTC(),
		Data: map[string]interface{}{
			"test": 1,
		},
	}
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(&r)
	_, err := BuildRequest(&buf)
	if err != nil {
		t.Errorf("Error during request validation: %s", err)
	}
}

func TestValidateRequest_Empty(t *testing.T) {
	var buf bytes.Buffer
	_, err := BuildRequest(&buf)
	if err == nil {
		t.Errorf("No error during empty request validation: %s", err)
	}
}

func TestValidateRequest_Bad(t *testing.T) {
	var buf bytes.Buffer
	r := Request{
		SchemaVersion: "0.1",
		Timestamp:     time.Now().UTC(),
		Data: map[string]interface{}{
			"test": 1,
		},
	}
	json.NewEncoder(&buf).Encode(&r)
	_, err := BuildRequest(&buf)
	if err == nil {
		t.Errorf("No error during empty request validation: %s", err)
	}
}
