package tracing

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/sdk/metric"
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

	otelProm := otelprom.New()
	provider := metric.NewMeterProvider(metric.WithReader(otelProm))
	registry := prometheus.NewRegistry()
	err = registry.Register(otelProm.Collector)
	if err != nil {
		logger.Panic("failed to initialize prometheus exporter", zap.Error(err))
	}

	global.SetMeterProvider(provider)
	otel.SetTracerProvider(tp)
	b3 := b3.New()
	otel.SetTextMapPropagator(b3)
	http.Handle("/", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	go func() {
		_ = http.ListenAndServe(":2222", nil) // nolint:gosec // we expect don't expose this interface to the internet
	}()

	logger.Info("Prometheus server running on :2222")
	return tp
}
