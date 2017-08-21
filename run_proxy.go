package main

import (
	"net/http"
	"net/http/httputil"
	"os"
	"time"

	"github.com/gorilla/mux"
)

type RunProxy struct {
	gen *RunGeneric
}

func (rp *RunProxy) Run() error {
	go rp.gen.Run()
	err := os.Chdir(args.ResourceDir)
	if err != nil {
		return err
	}
	proxy := &httputil.ReverseProxy{Director: director}
	r := mux.NewRouter()
	r.Handle("/{version}/{namespace}/projects/{projectID}/servers/{serverID}/endpoint/{service}{path:.*}", proxy)
	server := &http.Server{
		Addr:           ":8000",
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: http.DefaultMaxHeaderBytes,
	}
	return server.ListenAndServe()
}

func director(r *http.Request) {
	vars := mux.Vars(r)
	r.URL.Host = "localhost:8888"
	r.URL.Scheme = "http"
	r.URL.Path = vars["path"]
}
