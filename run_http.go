package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"regexp"

	"time"

	"github.com/gorilla/handlers"
)

const requestTimeout = 30 * time.Second

type RunHTTP struct{}

func (rh *RunHTTP) Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", ScriptHandler)
	server := &http.Server{
		Addr:        ":8000",
		ReadTimeout: 10 * time.Second,
		Handler:     handlers.LoggingHandler(os.Stdout, mux),
	}
	return server.ListenAndServe()
}

func ScriptHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	if r.Method != "POST" {
		appErr := AppError{StatusCode: http.StatusMethodNotAllowed}
		appErr.Write(ctx, w)
		return
	}
	if !checkToken(args.ApiRoot, r.Header.Get("Authorization")) {
		appErr := AppError{
			StatusCode: http.StatusForbidden,
		}
		appErr.Write(ctx, w)
		return
	}
	var requestData *Request
	var err error
	requestData, err = BuildRequest(r.Body)
	if err != nil {
		appErr := AppError{Err: err, StatusCode: http.StatusBadRequest}
		appErr.Write(ctx, w)
		return
	}
	resp := CreateResponseFromRequest(requestData)
	stats := NewStats()
	re := regexp.MustCompile(`\r?\n`)
	code := re.ReplaceAllString(fmt.Sprintf(args.Code, requestData), "")
	data, err := Run(ctx, stats, code)
	if err != nil {
		appErr := AppError{
			Err:        err,
			StatusCode: http.StatusBadRequest,
			Stacktrace: data,
		}
		appErr.Write(ctx, w)
		return
	}
	resp.Status = "ok"
	resp.ExecutionTime = stats.Duration()
	resp.Data = []byte(data)
	resp.Write(ctx, w)
}
