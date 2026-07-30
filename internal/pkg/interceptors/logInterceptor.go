package interceptors

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ctxLoggerKey struct{}

var LoggerCtxKey = ctxLoggerKey{}

var (
	grpcRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "notes",
			Name:      "grpc_requests_total",
			Help:      "Total number of gRPC requests.",
		},
		[]string{"rpc", "code"},
	)

	grpcDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "notes",
			Name:      "grpc_request_duration_seconds",
			Help:      "gRPC request duration in seconds.",
		},
		[]string{"rpc", "code"},
	)
)

func LogInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		requestID := requestIDFromMetadata(ctx)
		if requestID == "" {
			requestID = uuid.NewString()
		}

		sublogger := log.With().
			Str("requestID", requestID).
			Str("rpc", info.FullMethod).
			Logger()

		ctx = context.WithValue(ctx, LoggerCtxKey, &sublogger)

		start := time.Now()
		sublogger.Info().Msg("in")
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		code := status.Code(err)

		event := sublogger.Info()
		if code == codes.Internal || code == codes.Unavailable {
			event = sublogger.Error()
		}

		event.
			Err(err).
			Str("grpc_code", code.String()).
			Int64("time_ms", duration.Milliseconds()).
			Msg("out")

		grpcCode := status.Code(err).String()

		grpcRequests.
			WithLabelValues(info.FullMethod, grpcCode).
			Inc()

		grpcDuration.
			WithLabelValues(info.FullMethod, grpcCode).
			Observe(duration.Seconds())

		return resp, err
	}
}

func requestIDFromMetadata(ctx context.Context) string {
	values := metadata.ValueFromIncomingContext(ctx, "x-request-id")
	if len(values) == 0 {
		return ""
	}

	return values[0]
}
