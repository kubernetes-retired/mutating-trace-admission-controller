package main

import (
	"context"
	"encoding/base64"
	"errors"

	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
)

// TraceConfig is the struct used to pass configuration to the tracer
type TraceConfig struct {
	samplingRate float64
}

// ConfigureTracing will take passed configuration and set the sampling policy accordingly
func ConfigureTracing(config *TraceConfig) error {
	//validate configuration
	if config.samplingRate < 0.0 || config.samplingRate > 1.0 {
		return errors.New("invalid sample rate: must be between 0 and 1 inclusive")
	}

	trace.ApplyConfig(trace.Config{DefaultSampler: trace.ProbabilitySampler(config.samplingRate)})
	return nil
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
