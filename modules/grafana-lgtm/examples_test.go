package grafanalgtm_test

import (
	"context"
	"errors"
	"fmt"
	golog "log"
	"log/slog"
	"math/rand"
	"sync"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/log/global"
	metricsapi "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/testcontainers/testcontainers-go/modules/grafanalgtm"
)

func ExampleRun() {
	// runGrafanaLGTMContainer {
	ctx := context.Background()

	grafanaLgtmContainer, err := grafanalgtm.Run(ctx, "grafana/otel-lgtm:0.6.0")
	if err != nil {
		golog.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := grafanaLgtmContainer.Terminate(ctx); err != nil {
			golog.Fatalf("failed to terminate container: %s", err) // nolint:gocritic
		}
	}()
	// }

	state, err := grafanaLgtmContainer.State(ctx)
	if err != nil {
		golog.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRun_otelCollector() {
	ctx := context.Background()

	ctr, err := grafanalgtm.Run(ctx, "grafana/otel-lgtm:0.6.0", grafanalgtm.WithAdminCredentials("admin", "123456789"))
	if err != nil {
		golog.Fatalf("failed to start Grafana LGTM container: %s", err)
	}
	defer func() {
		if err := ctr.Terminate(ctx); err != nil {
			golog.Fatalf("failed to terminate Grafana LGTM container: %s", err)
		}
	}()

	// Set up OpenTelemetry.
	otelShutdown, err := setupOTelSDK(ctx, ctr)
	if err != nil {
		return
	}
	// Handle shutdown properly so nothing leaks.
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	// roll dice 10000 times, concurrently
	max := 10_000
	wg := sync.WaitGroup{}
	for i := 0; i < max; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			rolldice(ctx)
		}()
	}

	wg.Wait()

	// Output:
	// shutdown errors: <nil>
}

// setupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func setupOTelSDK(ctx context.Context, ctr *grafanalgtm.GrafanaLGTMContainer) (shutdown func(context.Context) error, err error) { // nolint:nonamedreturns // this is a pattern in the OpenTelemetry Go SDK
	var shutdownFuncs []func(context.Context) error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		fmt.Println("shutdown errors:", err)
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	prop := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	otel.SetTextMapPropagator(prop)

	otlpHttpEndpoint := ctr.MustOtlpHttpEndpoint(ctx)

	traceExporter, err := otlptrace.New(ctx,
		otlptracehttp.NewClient(
			// adding schema to avoid this error:
			// 2024/07/19 13:16:30 internal_logging.go:50: "msg"="otlptrace: parse endpoint url" "error"="parse \"127.0.0.1:33007\": first path segment in URL cannot contain colon" "url"="127.0.0.1:33007"
			// it does not happen with the logs and metrics exporters
			otlptracehttp.WithEndpointURL("http://"+otlpHttpEndpoint),
			otlptracehttp.WithInsecure(),
		),
	)
	if err != nil {
		return nil, err
	}

	tracerProvider := trace.NewTracerProvider(trace.WithBatcher(traceExporter))
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	metricExporter, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithInsecure(),
		otlpmetrichttp.WithEndpoint(otlpHttpEndpoint),
	)
	if err != nil {
		return nil, err
	}

	// The exporter embeds a default OpenTelemetry Reader and
	// implements prometheus.Collector, allowing it to be used as
	// both a Reader and Collector.
	prometheusExporter, err := prometheus.New()
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter)),
		metric.WithReader(prometheusExporter),
	)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	logExporter, err := otlploghttp.New(ctx,
		otlploghttp.WithInsecure(),
		otlploghttp.WithEndpoint(otlpHttpEndpoint),
	)
	if err != nil {
		return nil, err
	}

	loggerProvider := log.NewLoggerProvider(log.WithProcessor(log.NewBatchProcessor(logExporter)))
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
	global.SetLoggerProvider(loggerProvider)

	err = runtime.Start(runtime.WithMinimumReadMemStatsInterval(time.Second))
	if err != nil {
		logger.ErrorContext(ctx, "otel runtime instrumentation failed:", err) // nolint:all // this is a pattern in the OpenTelemetry Go SDK
	}

	return
}

// rollDiceApp {
const schemaName = "https://github.com/grafana/docker-otel-lgtm"

var (
	tracer = otel.Tracer(schemaName)
	logger = otelslog.NewLogger(schemaName)
	meter  = otel.Meter(schemaName)
)

func rolldice(ctx context.Context) {
	ctx, span := tracer.Start(ctx, "roll")
	defer span.End()

	// 20-sided dice
	roll := 1 + rand.Intn(20)
	logger.InfoContext(ctx, fmt.Sprintf("Rolled a dice: %d\n", roll), slog.Int("result", roll))

	opt := metricsapi.WithAttributes(
		attribute.Key("sides").Int(roll),
	)

	// This is the equivalent of prometheus.NewCounterVec
	counter, err := meter.Int64Counter("rolldice-counter", metricsapi.WithDescription("a 20-sided dice"))
	if err != nil {
		golog.Fatal(err)
	}
	counter.Add(ctx, int64(roll), opt)
}

// }
