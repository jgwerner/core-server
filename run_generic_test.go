package main

import (
	"bytes"
	"net/http/httptest"
	"os"
	"testing"
)

func TestGenericRun(t *testing.T) {
	var ts *httptest.Server
	ts, args = mockAPI(`{}`)
	args.ResourceDir = "/tmp"
	defer ts.Close()
	var buf bytes.Buffer
	out = &buf
	defer func() { out = os.Stderr }()
	rg := &RunGeneric{
		command: "echo",
		args:    []string{"test"},
	}
	err := rg.Run()
	if err != nil {
		t.Error(err)
	}
}
