package main

import (
	"fmt"
	"net/http"
	"io"
	"os"
	"go.opencensus.io/exporter/stackdriver"
	"go.opencensus.io/trace"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/plugin/ochttp"
	"cloud.google.com/go/logging"
	"cloud.google.com/go/profiler"
	"encoding/json"
	"runtime/debug"
	"io/ioutil"
	_ "golang.org/x/sync/errgroup"
	"golang.org/x/sync/errgroup"
)


func main() {
	if err := profiler.Start(profiler.Config{ProjectID: "apstndb-sandbox", Service: "frontend", ServiceVersion: "1.0.0"}); err != nil {
		panic(err)
	}
	exporter, err := stackdriver.NewExporter(stackdriver.Options{})
	if err != nil {
		panic(err)
	}

	trace.RegisterExporter(exporter)
	view.RegisterExporter(exporter)
	trace.SetDefaultSampler(trace.AlwaysSample())

	client := &http.Client{Transport: &ochttp.Transport{}}

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)

		errgrp := errgroup.Group{}
		errgrp.Go(func() error {
			{
				r, _ := http.NewRequest("GET", "http://backend", nil)
				resp, _ := client.Do(r.WithContext(request.Context()))
				bytes, _ := ioutil.ReadAll(resp.Body)
				fmt.Fprint(io.MultiWriter(writer, os.Stdout), string(bytes))
			}
			return nil
		})
		errgrp.Go(func() error {
			{
				r, _ := http.NewRequest("GET", "http://backend", nil)
				resp, _ := client.Do(r.WithContext(request.Context()))
				bytes, _ := ioutil.ReadAll(resp.Body)
				fmt.Fprintln(io.MultiWriter(writer, os.Stdout), string(bytes))
			}
			return nil
		})
		errgrp.Wait()
	})
	http.HandleFunc("/healthz", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/error", func(writer http.ResponseWriter, request *http.Request) {
		serviceName := "frontend"
		stackTrace := string(debug.Stack())
		traceID := trace.FromContext(request.Context()).SpanContext().TraceID

		entry := debugEntry(serviceName, stackTrace, request, traceID)
		json.NewEncoder(os.Stderr).Encode(entry)
		// json.RawMessage(fmt.Sprintf(`{"serviceContext": {"service": "manual-testing"}, "message": "Test Error", "context": {"httpRequest": {"url": "/debug", "method": "GET", "responseStatusCode": 500}}}`))})

		writer.WriteHeader(http.StatusInternalServerError)
	})

	fmt.Fprintln(os.Stderr, "Started")
	http.ListenAndServe(":8080", &ochttp.Handler{})
}

func debugEntry(serviceName string, stackTrace string, request *http.Request, traceID trace.TraceID) logging.Entry {
	payload := map[string]interface{}{
		"serviceContext": map[string]interface{}{"service": serviceName},
		"message":        stackTrace,
		"context": map[string]interface{}{
			"httpRequest": map[string]interface{}{
				"url": request.URL.Path, "method": request.Method,
			}}}
	entry := logging.Entry{Severity: logging.Error,
		Trace: traceID.String(),
		Payload: payload,
	}
	return entry
}
