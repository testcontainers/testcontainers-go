package grafanalgtm_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/testcontainers/testcontainers-go/modules/grafanalgtm"
)

func ExampleRun() {
	// runGrafanaLGTMContainer {
	ctx := context.Background()

	grafanaLgtmContainer, err := grafanalgtm.Run(ctx, "grafana/otel-lgtm:0.6.0")
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := grafanaLgtmContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err) // nolint:gocritic
		}
	}()
	// }

	state, err := grafanaLgtmContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRun_otelCollector() {
	ctx := context.Background()

	ctr, err := grafanalgtm.Run(ctx, "grafana/otel-lgtm:0.6.0")
	if err != nil {
		log.Fatalf("failed to start Grafana LGTM container: %s", err)
	}
	defer func() {
		if err := ctr.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate Grafana LGTM container: %s", err)
		}
	}()

	// start otel tracing
	if shutdown := retryInitTracer(ctr.MustOtlpGrpcURL(ctx)); shutdown != nil {
		defer shutdown()
	}

	// add custom attributes and events to the span
	_, span := otel.Tracer("GetServiceDetail").Start(ctx,
		"spanMetricDao.GetServiceDetail",
		trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	span.SetAttributes(attribute.String("controller", "books"))
	span.AddEvent("event")

	req, err := http.NewRequest(http.MethodGet, ctr.MustHttpURL(context.Background()), nil)
	if err != nil {
		log.Fatalf("failed to create request: %s", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("failed to send request: %s", err)
	}
	defer res.Body.Close()

	fmt.Println(res.StatusCode)

	// Output:
	// 200
}

var tracerExp *otlptrace.Exporter

func retryInitTracer(otlpEndpoint string) func() {
	var shutdown func()
	go func() {
		for {
			// otel will reconnected and re-send spans when otel col recover. so, we don't need to re-init tracer exporter.
			if tracerExp == nil {
				shutdown = initTracer(otlpEndpoint)
			} else {
				break
			}
			time.Sleep(time.Minute * 5)
		}
	}()

	return shutdown
}

func initTracer(otlpEndpoint string) func() {
	// temporarily set timeout to 10s
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	serviceName := "service_name"
	os.Setenv("OTEL_SERVICE_NAME", serviceName)

	otelAgentAddr := otlpEndpoint
	os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", otelAgentAddr)

	log.Printf("OTLP Trace connect to: %s with service name: %s", otelAgentAddr, serviceName)

	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithInsecure(), otlptracegrpc.WithDialOption())
	if err != nil {
		handleErr(err, "OTLP Trace gRPC Creation")
		return nil
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(resource.NewWithAttributes(semconv.SchemaURL)))

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	tracerExp = traceExporter
	return func() {
		// Shutdown will flush any remaining spans and shut down the exporter.
		handleErr(tracerProvider.Shutdown(ctx), "failed to shutdown TracerProvider")
	}
}

func handleErr(err error, message string) {
	if err != nil {
		log.Fatalf("%s: %v", message, err)
	}
}
