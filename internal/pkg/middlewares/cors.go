package middlewares

import (
	"fmt"
	"net/http"

	"github.com/rs/zerolog"
)

func CORS(next http.Handler, allowedOrigin string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		var logger zerolog.Logger
		ctx := r.Context()
		logger, ok := ctx.Value(LoggerCtxKey).(zerolog.Logger)
		if !ok {
			logger = zerolog.Nop()
		}

		if origin == allowedOrigin {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Add("Vary", "Origin")
			w.Header().Set(
				"Access-Control-Allow-Methods",
				"GET, POST, OPTIONS",
			)
			w.Header().Set(
				"Access-Control-Allow-Headers",
				"Authorization, Content-Type, X-Request-ID",
			)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if r.Method == http.MethodOptions {
			if origin != allowedOrigin {
				logger.Error().Err(fmt.Errorf("origin is now allowed: %q", origin)).Msg("cors mw")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"origin is not allowed"}`))
				return
			}

			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
