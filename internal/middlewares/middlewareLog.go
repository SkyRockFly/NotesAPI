package middlewares

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type ctxLoggerKey struct{}

var LoggerCtxKey = ctxLoggerKey{}

type middlewareResponseWriter struct {
	w          http.ResponseWriter
	statusCode int
}

func (r *middlewareResponseWriter) Header() http.Header {
	return r.w.Header()
}

func (r *middlewareResponseWriter) Write(w []byte) (int, error) {
	return r.w.Write(w)
}

func (r *middlewareResponseWriter) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.w.WriteHeader(statusCode)
}

func LogMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("x-request-id")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		sw := &middlewareResponseWriter{w: w, statusCode: -1}

		subLogger := log.With().Str("requestID", requestID).Logger()

		subLogger.Info().
			Str("path", r.URL.Path).
			Str("method", r.Method).Msg("in")

		ctx := context.WithValue(r.Context(), LoggerCtxKey, subLogger)
		next(sw, r.WithContext(ctx))
		subLogger.Info().Int("status", sw.statusCode).Msg("out")
	}
}
