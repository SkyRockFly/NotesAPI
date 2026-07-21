package interceptors

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ctxLoggerKey struct{}

var LoggerCtxKey = ctxLoggerKey{}

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

		code := status.Code(err)

		event := sublogger.Info()
		if code == codes.Internal || code == codes.Unavailable {
			event = sublogger.Error()
		}

		event.
			Err(err).
			Str("grpc_code", code.String()).
			Int64("time_ms", time.Since(start).Milliseconds()).
			Msg("out")

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
