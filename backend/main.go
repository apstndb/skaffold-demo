package main

import (
	"net/http"
	"go.opencensus.io/exporter/stackdriver"
	"go.opencensus.io/trace"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/plugin/ochttp"
	"fmt"
)


func main() {
	exporter, err := stackdriver.NewExporter(stackdriver.Options{})
	if err != nil {
		panic(err)
	}

	trace.RegisterExporter(exporter)
	view.RegisterExporter(exporter)
	trace.SetDefaultSampler(trace.AlwaysSample())

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		fmt.Fprintln(writer, "Hello from backend")
	})

	http.HandleFunc("/healthz", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})

	fmt.Println("Started")

	http.ListenAndServe(":8080", &ochttp.Handler{})
}
