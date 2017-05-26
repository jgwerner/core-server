package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
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
