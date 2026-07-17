package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"
)

type userIDCtx struct{}

var UIDKey = userIDCtx{}

func Auth(secret []byte, leeway time.Duration) func(http.HandlerFunc) http.HandlerFunc {
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithLeeway(leeway))

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")

			var logger zerolog.Logger
			ctx := r.Context()
			logger, ok := ctx.Value(LoggerCtxKey).(zerolog.Logger)
			if !ok {
				logger = zerolog.Nop()
			}

			parts := strings.SplitN(auth, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				logger.Info().Err(fmt.Errorf("bad auth header: %s", auth)).Msg("auth middleware")
				http.Error(w, `"error":"invalid auth header"`, http.StatusBadRequest)
				return
			}

			rawKey := strings.TrimSpace(parts[1])

			var claims jwt.RegisteredClaims
			if _, err := parser.ParseWithClaims(rawKey, &claims, func(*jwt.Token) (any, error) {
				return secret, nil
			}); err != nil {
				logger.Error().Err(fmt.Errorf("parse jwt: %w", err)).Msg("auth middleware")
				http.Error(w, `"error":"unauthorized"`, http.StatusUnauthorized)
				return
			}

			uid, err := strconv.Atoi(claims.Subject)
			if err != nil {
				logger.Error().Err(fmt.Errorf("convert uid: %w", err)).Msg("auth middleware")
				http.Error(w, `"error":"service error"`, http.StatusInternalServerError)
				return
			}
			idCtx := context.WithValue(ctx, UIDKey, uid)

			next.ServeHTTP(w, r.WithContext(idCtx))
		}
	}
}
