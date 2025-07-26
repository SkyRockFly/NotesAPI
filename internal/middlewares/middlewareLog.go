package middlewares

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const (
	LoggerCtxKey = iota
)

type APIResponse struct {
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:",omitempty"`
}

type HandlerFuncWithStatus func(writer http.ResponseWriter, request *http.Request) (APIResponse, int, error)

func LogMiddleware(next HandlerFuncWithStatus) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("x-request-id")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		subLogger := log.With().Str("requestID", requestID).Logger()

		subLogger.Info().
			Str("path", r.URL.Path).
			Str("method", r.Method).Msg("in")

		ctx := context.WithValue(r.Context(), LoggerCtxKey, subLogger)
		data, statusCode, err := next(w, r.WithContext(ctx))
		if err != nil {
			http.Error(w, err.Error(), statusCode)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(data.Data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		subLogger.Info().Int("status", statusCode).Msg("out")
	}
}
