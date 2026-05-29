// Package interceptor provides gRPC server and client interceptors that report
// errors and panics to Sentry, propagate distributed traces via gRPC metadata,
// and start Sentry transactions/spans for every RPC call.
package interceptor

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Option configures the Sentry interceptor behaviour.
type Option interface {
	apply(*options)
}

type options struct {
	repanic         bool
	waitForDelivery bool
	timeout         time.Duration
	reportOn        func(error) bool
}

var defaultOptions = options{
	repanic:         false,
	waitForDelivery: false,
	timeout:         2 * time.Second,
	reportOn:        func(error) bool { return true },
}

func buildOptions(opts []Option) options {
	o := defaultOptions
	for _, opt := range opts {
		opt.apply(&o)
	}
	return o
}

type repanicOption bool

func (r repanicOption) apply(o *options) { o.repanic = bool(r) }

// WithRepanic configures whether the interceptor should re-panic after
// recovering and reporting a panic to Sentry.
func WithRepanic(b bool) Option { return repanicOption(b) }

type waitForDeliveryOption bool

func (w waitForDeliveryOption) apply(o *options) { o.waitForDelivery = bool(w) }

// WithWaitForDelivery blocks the RPC response until Sentry has delivered the
// event (up to the configured timeout).
func WithWaitForDelivery(b bool) Option { return waitForDeliveryOption(b) }

type timeoutOption time.Duration

func (t timeoutOption) apply(o *options) { o.timeout = time.Duration(t) }

// WithTimeout sets the maximum time to wait for Sentry event delivery.
func WithTimeout(t time.Duration) Option { return timeoutOption(t) }

type reportOnOption func(error) bool

func (r reportOnOption) apply(o *options) { o.reportOn = r }

// WithReportOn overrides the predicate used to decide whether an error should
// be reported to Sentry. Defaults to always reporting.
func WithReportOn(f func(error) bool) Option { return reportOnOption(f) }

// ReportOnCodes returns a predicate that only reports errors whose gRPC status
// code is in cc.
func ReportOnCodes(cc ...codes.Code) func(error) bool {
	return func(err error) bool {
		for _, c := range cc {
			if status.Code(err) == c {
				return true
			}
		}
		return false
	}
}

// wrappedStream carries a custom context through a server stream.
type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context { return w.ctx }

// recoverWithSentry reports a panic to Sentry and optionally re-panics.
func recoverWithSentry(hub *sentry.Hub, ctx context.Context, o options) {
	if err := recover(); err != nil {
		hub.RecoverWithContext(ctx, err)
		if o.waitForDelivery {
			hub.Flush(o.timeout)
		}
		if o.repanic {
			panic(err)
		}
	}
}

// hubFromContext returns a per-request hub, creating one if necessary.
func hubFromContext(ctx context.Context) (*sentry.Hub, context.Context) {
	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub().Clone()
		ctx = sentry.SetHubOnContext(ctx, hub)
	}
	return hub, ctx
}

// serverTransactionOpts builds the SpanOption slice for StartTransaction.
// continueFromMetadata may return nil when there is no incoming trace header;
// passing a nil SpanOption to the Sentry SDK causes a nil-pointer panic because
// the SDK calls every option unconditionally, so we filter it out here.
func serverTransactionOpts(method string, md metadata.MD) []sentry.SpanOption {
	opts := []sentry.SpanOption{
		sentry.WithOpName("grpc.server"),
		sentry.WithDescription(method),
		sentry.WithTransactionSource(sentry.SourceURL),
	}
	if traceOpt := continueFromMetadata(md); traceOpt != nil {
		opts = append(opts, traceOpt)
	}
	return opts
}

// continueFromMetadata extracts a Sentry trace from incoming gRPC metadata.
// Returns nil when no trace header is present.
func continueFromMetadata(md metadata.MD) sentry.SpanOption {
	if md == nil {
		return nil
	}
	var trace, baggage string
	if v, ok := md[sentry.SentryTraceHeader]; ok && len(v) > 0 {
		trace = v[0]
	}
	if v, ok := md[sentry.SentryBaggageHeader]; ok && len(v) > 0 {
		baggage = v[0]
	}
	if trace == "" {
		return nil
	}
	return sentry.ContinueFromHeaders(trace, baggage)
}

func toSpanStatus(code codes.Code) sentry.SpanStatus {
	switch code {
	case codes.OK:
		return sentry.SpanStatusOK
	case codes.Canceled:
		return sentry.SpanStatusCanceled
	case codes.Unknown:
		return sentry.SpanStatusUnknown
	case codes.InvalidArgument:
		return sentry.SpanStatusInvalidArgument
	case codes.DeadlineExceeded:
		return sentry.SpanStatusDeadlineExceeded
	case codes.NotFound:
		return sentry.SpanStatusNotFound
	case codes.AlreadyExists:
		return sentry.SpanStatusAlreadyExists
	case codes.PermissionDenied:
		return sentry.SpanStatusPermissionDenied
	case codes.ResourceExhausted:
		return sentry.SpanStatusResourceExhausted
	case codes.FailedPrecondition:
		return sentry.SpanStatusFailedPrecondition
	case codes.Aborted:
		return sentry.SpanStatusAborted
	case codes.OutOfRange:
		return sentry.SpanStatusOutOfRange
	case codes.Unimplemented:
		return sentry.SpanStatusUnimplemented
	case codes.Internal:
		return sentry.SpanStatusInternalError
	case codes.Unavailable:
		return sentry.SpanStatusUnavailable
	case codes.DataLoss:
		return sentry.SpanStatusDataLoss
	case codes.Unauthenticated:
		return sentry.SpanStatusUnauthenticated
	default:
		return sentry.SpanStatusUndefined
	}
}

// UnaryServerInterceptor returns a gRPC unary server interceptor that:
//   - clones a Sentry hub per request and attaches it to the context
//   - starts a Sentry transaction continuing any distributed trace present in
//     the incoming gRPC metadata
//   - captures panics and errors via Sentry
func UnaryServerInterceptor(opts ...Option) grpc.UnaryServerInterceptor {
	o := buildOptions(opts)
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		hub, ctx := hubFromContext(ctx)

		md, _ := metadata.FromIncomingContext(ctx)
		tx := sentry.StartTransaction(ctx, info.FullMethod, serverTransactionOpts(info.FullMethod, md)...)
		tx.SetData("grpc.request.method", info.FullMethod)
		ctx = tx.Context()
		defer tx.Finish()

		defer recoverWithSentry(hub, ctx, o)

		resp, err = handler(ctx, req)
		if err != nil && o.reportOn(err) {
			hub.CaptureException(err)
			tx.Sampled = sentry.SampledTrue
		}
		tx.Status = toSpanStatus(status.Code(err))
		return resp, err
	}
}

// StreamServerInterceptor returns a gRPC stream server interceptor with the
// same Sentry behaviour as UnaryServerInterceptor.
func StreamServerInterceptor(opts ...Option) grpc.StreamServerInterceptor {
	o := buildOptions(opts)
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		ctx := ss.Context()
		hub, ctx := hubFromContext(ctx)

		md, _ := metadata.FromIncomingContext(ctx)
		tx := sentry.StartTransaction(ctx, info.FullMethod, serverTransactionOpts(info.FullMethod, md)...)
		tx.SetData("grpc.request.method", info.FullMethod)
		ctx = tx.Context()
		defer tx.Finish()

		defer recoverWithSentry(hub, ctx, o)

		err = handler(srv, &wrappedStream{ss, ctx})
		if err != nil && o.reportOn(err) {
			hub.CaptureException(err)
			tx.Sampled = sentry.SampledTrue
		}
		tx.Status = toSpanStatus(status.Code(err))
		return err
	}
}

// UnaryClientInterceptor returns a gRPC unary client interceptor that starts a
// Sentry child span and propagates the trace to the remote service via gRPC
// metadata.
func UnaryClientInterceptor(opts ...Option) grpc.UnaryClientInterceptor {
	o := buildOptions(opts)
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, callOpts ...grpc.CallOption) error {
		hub, ctx := hubFromContext(ctx)

		span := sentry.StartSpan(ctx, "grpc.client", sentry.WithDescription(method))
		span.SetData("grpc.request.method", method)
		ctx = span.Context()
		ctx = injectSentryTrace(ctx, span)
		defer span.Finish()

		err := invoker(ctx, method, req, reply, cc, callOpts...)
		if err != nil && o.reportOn(err) {
			hub.CaptureException(err)
		}
		return err
	}
}

// StreamClientInterceptor returns a gRPC stream client interceptor that starts
// a Sentry child span and propagates the trace to the remote service.
func StreamClientInterceptor(opts ...Option) grpc.StreamClientInterceptor {
	o := buildOptions(opts)
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, callOpts ...grpc.CallOption) (grpc.ClientStream, error) {
		hub, ctx := hubFromContext(ctx)

		span := sentry.StartSpan(ctx, "grpc.client", sentry.WithDescription(method))
		span.SetData("grpc.request.method", method)
		ctx = span.Context()
		ctx = injectSentryTrace(ctx, span)
		defer span.Finish()

		cs, err := streamer(ctx, desc, cc, method, callOpts...)
		if err != nil && o.reportOn(err) {
			hub.CaptureException(err)
		}
		return cs, err
	}
}

// injectSentryTrace appends sentry-trace and baggage headers to the outgoing
// gRPC metadata so downstream services can continue the trace.
func injectSentryTrace(ctx context.Context, span *sentry.Span) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if ok {
		md = md.Copy()
	} else {
		md = metadata.MD{}
	}
	md.Set(sentry.SentryTraceHeader, span.ToSentryTrace())
	md.Set(sentry.SentryBaggageHeader, span.ToBaggage())
	return metadata.NewOutgoingContext(ctx, md)
}
