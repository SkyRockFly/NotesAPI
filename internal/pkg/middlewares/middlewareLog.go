package middlewares

import (
	"context"
	"fmt"
	"net/http"
	requestid "notes/internal/pkg/requestID"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
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
	num, err := r.w.Write(w)
	if err != nil {
		return 0, fmt.Errorf("%w", err)
	}
	return num, nil
}

func (r *middlewareResponseWriter) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.w.WriteHeader(statusCode)
}

var (
	httpRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "gateway",
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests.",
		},
		[]string{"method", "route", "status"},
	)

	httpDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "gateway",
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration in seconds.",
		},
		[]string{"method", "route", "status"},
	)
)

func LogMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("x-request-id")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		ctx := requestid.WithContext(
			r.Context(),
			requestID,
		)

		route := r.Pattern
		if route == "" {
			route = r.Method + " " + r.URL.Path
		}

		sw := &middlewareResponseWriter{w: w, statusCode: -1}

		subLogger := log.With().
			Str("requestID", requestID).
			Logger()

		subLogger.Info().
			Str("method", r.Method).
			Str("route", route).Msg("in")

		ctx = context.WithValue(
			ctx,
			LoggerCtxKey,
			subLogger,
		)

		startTime := time.Now()
		next(sw, r.WithContext(ctx))
		duration := time.Since(startTime)

		event := subLogger.Info()

		switch {
		case sw.statusCode >= 500:
			event = subLogger.Error()
		case sw.statusCode >= 400:
			event = subLogger.Warn()
		}

		event.
			Int("status", sw.statusCode).
			Int64("time_ms", duration.Milliseconds()).
			Msg("out")

		statusText := strconv.Itoa(sw.statusCode)
		httpRequests.WithLabelValues(r.Method, route, statusText).Inc()
		httpDuration.WithLabelValues(r.Method, route, statusText).Observe(duration.Seconds())
	}
}
