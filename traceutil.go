package main

import (
	"context"
	"encoding/base64"

	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
)

// ConfigureTracing will take passed configuration and set the sampling policy accordingly
func ConfigureTracing() {
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
}

// GenerateEmbeddableSpanContext takes a SpanContext and returns a serialized string
func GenerateEmbeddableSpanContext() string {
	// should not be exported, purpose of this span is to retrieve OC compliant SpanContext
	_, tempSpan := trace.StartSpan(context.Background(), "")
	tempSpanContext := tempSpan.SpanContext()

	rawContextBytes := propagation.Binary(tempSpanContext)
	encodedContext := base64.StdEncoding.EncodeToString(rawContextBytes)

	return encodedContext
}
