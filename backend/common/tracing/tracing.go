package tracing

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric/global"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	"go.opentelemetry.io/otel/sdk/metric/export/aggregation"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	selector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// NewGRPUnaryClientInterceptor returns unary client interceptor. It is used with `grpc.WithUnaryInterceptor` method.
func NewGRPUnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return otelgrpc.UnaryClientInterceptor()
}

// NewGRPUnaryServerInterceptor returns unary server interceptor. It is used with `grpc.UnaryInterceptor` method.
func NewGRPUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return otelgrpc.UnaryServerInterceptor()
}

// NewGRPCStreamClientInterceptor returns stream client interceptor. It is used with `grpc.WithStreamInterceptor` method.
func NewGRPCStreamClientInterceptor() grpc.StreamClientInterceptor {
	return otelgrpc.StreamClientInterceptor()
}

// NewGRPCStreamServerInterceptor returns stream server interceptor. It is used with `grpc.StreamInterceptor` method.
func NewGRPCStreamServerInterceptor() grpc.StreamServerInterceptor {
	return otelgrpc.StreamServerInterceptor()
}

// InitTracer performs the initialization of the traceprovider.
// By default this tries to init a jeager tracer.
func InitTracer(logger *zap.Logger, name string) *sdktrace.TracerProvider {
	//exporter, err := stdout.New(stdout.WithPrettyPrint())
	logger.Debug("Setting up tracing provider")
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint())

	if err != nil {
		logger.Fatal("Error creating collector.", zap.Error(err))
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(name),
			)),
	)
	config := prometheus.Config{}
	c := controller.New(
		processor.NewFactory(
			selector.NewWithHistogramDistribution(),
			aggregation.CumulativeTemporalitySelector(),
			processor.WithMemory(false),
		),
		controller.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(name),
			),
		),
	)
	prom, err := prometheus.New(config, c)

	if err != nil {
		logger.Panic("failed to initialize prometheus exporter", zap.Error(err))
	}

	global.SetMeterProvider(prom.MeterProvider())

	otel.SetTracerProvider(tp)
	b3 := b3.New()
	otel.SetTextMapPropagator(b3)
	http.HandleFunc("/", prom.ServeHTTP)
	go func() {
		_ = http.ListenAndServe(":2222", nil)
	}()

	logger.Info("Prometheus server running on :2222")
	return tp
}
