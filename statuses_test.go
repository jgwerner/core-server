package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

func mockAPIServer(h http.Handler) (*httptest.Server, *Args) {
	ts := httptest.NewServer(h)
	apiRoot, _ := url.Parse(ts.URL)
	args := &Args{
		ApiRoot: fmt.Sprintf("%s:%s", apiRoot.Hostname(), apiRoot.Port()),
	}
	return ts, args
}

func mockAPIHandler(data io.Reader) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.Copy(w, data)
	})
}

func mockAPI(data string) (*httptest.Server, *Args) {
	return mockAPIServer(mockAPIHandler(bytes.NewBufferString(data)))
}

func TestSetStatus(t *testing.T) {
	ts, args := mockAPI(`{"status": "Stopped"}`)
	defer ts.Close()
	status := "Running"
	var buf bytes.Buffer
	log.SetOutput(&buf)
	SetStatus(args, status)
	log.SetOutput(os.Stderr)
	if buf.Len() > 0 {
		log.Println(buf.String())
		t.Error("Error during status change")
	}
}
