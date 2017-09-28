package main

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"time"

	"github.com/gorilla/handlers"
)

const requestTimeout = 30 * time.Second

type RunHTTP struct{}

func (rh *RunHTTP) Run() error {
	RunKernelGateway(out, out, args.KernelName)
	GetKernel()
	server := &http.Server{
		Addr:        ":6006",
		ReadTimeout: 10 * time.Second,
		Handler:     handlers.LoggingHandler(out, http.HandlerFunc(ScriptHandler)),
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
	if !checkToken(args.ApiRoot, r.URL.Query().Get("access_token")) {
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
	rawCode := fmt.Sprintf(`%s('%s')`, args.Function, requestData.DataString())
	code := re.ReplaceAllString(rawCode, "")
	data, err := Run(ctx, stats, args.Script, code)
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
