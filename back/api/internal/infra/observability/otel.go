package observability

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// InitTracer initializes an OTLP exporter, and configures the corresponding trace and
// metric providers.
func InitTracer(ctx context.Context, serviceName string) (func(context.Context) error, error) {
	var exporter sdktrace.SpanExporter
	var err error

	// If OTEL_EXPORTER_OTLP_ENDPOINT is set, use OTLP gRPC exporter.
	// Otherwise, fallback to stdout for testing/dev purposes.
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" {
		exporter, err = otlptracegrpc.New(ctx, otlptracegrpc.WithInsecure())
		if err != nil {
			return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
		}
	} else {
		// Use stdout as a fallback
		exporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, fmt.Errorf("failed to create stdout trace exporter: %w", err)
		}
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(exporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)

	var metricExporter sdkmetric.Exporter
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" {
		metricExporter, _ = otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithInsecure())
	}
	var meterProvider *sdkmetric.MeterProvider
	if metricExporter != nil {
		meterProvider = sdkmetric.NewMeterProvider(
			sdkmetric.WithResource(res),
			sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
		)
		otel.SetMeterProvider(meterProvider)
	}

	// Set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// return a shutdown function so caller can defer it
	return func(c context.Context) error {
		if meterProvider != nil {
			_ = meterProvider.Shutdown(c)
		}
		return tracerProvider.Shutdown(c)
	}, nil
}
