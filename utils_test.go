package core

import (
	"testing"
	"bytes"
	"encoding/json"
	"time"
	"github.com/speps/go-hashids"
)

func TestDecodeHashID(t *testing.T) {
	expected := 111
	hd := hashids.NewData()
	hd.MinLength = 8
	hd.Salt = "test-123-test"
	hashId, err := hashids.NewWithData(hd).Encode([]int{expected})
	if err != nil {
		t.Error(err)
	}
	actual, err := DecodeHashID(hd.Salt, hashId)
	if err != nil {
		t.Error(err)
	}
	if actual != expected {
		t.Error("Ids don't match")
	}
}

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
