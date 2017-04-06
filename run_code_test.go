package main

import (
	"net/http/httptest"
	"testing"
)

func TestCodeRun(t *testing.T) {
	var ts *httptest.Server
	ts, args = mockAPI(`{}`)
	defer ts.Close()
	prepareScript("def test():\n\treturn 'test'")
	rc := &RunCode{}
	err := rc.Run()
	if err != nil {
		t.Error(err)
	}
}
