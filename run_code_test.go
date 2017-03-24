package main

import (
	"net/http/httptest"
	"testing"
)

func TestCodeRun(t *testing.T) {
	var ts *httptest.Server
	ts, args = mockAPI(`{}`)
	defer ts.Close()
	rc := &RunCode{}
	args.Code = "print('test')"
	err := rc.Run()
	if err != nil {
		t.Error(err)
	}
}
