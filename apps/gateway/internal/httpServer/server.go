package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	usernoterepo "notes/apps/gateway/internal/userNoteRepo"
	userservice "notes/apps/gateway/internal/userService"
	"notes/pkg/middlewares"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

const (
	encodeError = `{"error":"encode failed"}`
)

var isShuttingDown atomic.Bool

type errorAPIResponse struct {
	Err string `json:"error"`
}

type healthResponse struct {
	Health bool `json:"health"`
}

type ResponseNote struct {
	ID        int       `json:"id"`
	AccountID int       `json:"account_id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func StartServer(ctx context.Context, port string, service *userservice.Service) error {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /note/get",
		middlewares.LogMiddleware(
			middlewares.JSONFileSizeMiddleware(
				getNoteHandler(service))))

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	errs, eCtx := errgroup.WithContext(ctx)
	errs.Go(func() error {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("listenAndServe: %w", err)
		}
		return nil
	})

	<-eCtx.Done()
	isShuttingDown.Store(true)

	shutdownCtx, done := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer done()

	if err := server.Shutdown(shutdownCtx); err != nil &&
		!errors.Is(err, http.ErrServerClosed) &&
		!errors.Is(err, context.Canceled) {
		if errors.Is(err, context.DeadlineExceeded) {
			_ = server.Close()
		}
		log.Warn().Err(err).Msg("graceful shutdown")
	}

	if err := errs.Wait(); err != nil {
		return fmt.Errorf("server: %w", err)
	}
	return nil
}

func HealthCheckHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := getCtxLogger(r.Context())
		if isShuttingDown.Load() {
			writeJSON(w, http.StatusServiceUnavailable, logger, errorAPIResponse{Err: "shutting down"})
			logger.
				Warn().
				Msg("httpServer.healthcheckHandler: shutting down")
			return
		}
		resp := &healthResponse{Health: true}
		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			logger.
				Warn().
				Err(err).
				Msg("httpServer.healthcheckHandler: json.Encode failed")
			return
		}
	}
}

func writeJSON(w http.ResponseWriter, statusCode int, logger *zerolog.Logger, resp any) {
	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(resp); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(encodeError))
		logger.Error().Err(fmt.Errorf("encode: %w", err)).Msg("write JSON")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if _, err := w.Write(buf.Bytes()); err != nil {
		logger.Error().Err(fmt.Errorf("write: %w", err)).Msg("write JSON")
		return
	}
}

func remapSvcToRespNote(note userservice.Note) ResponseNote {
	newNote := ResponseNote{
		ID:        note.ID,
		AccountID: note.AccountID,
		Title:     note.Title,
		Body:      note.Body,
		CreatedAt: note.CreatedAt,
		UpdatedAt: note.UpdatedAt,
	}

	return newNote
}

func getCtxLogger(ctx context.Context) *zerolog.Logger {
	if v := ctx.Value(middlewares.LoggerCtxKey); v != nil {
		if lg, ok := v.(zerolog.Logger); ok {
			return &lg
		}
	}
	logger := zerolog.Nop()
	return &logger
}

func mapToHTTPError(err error, log *zerolog.Logger) (int, string) {
	switch {
	case errors.Is(err, userservice.ErrInvalid):
		log.Info().Msg("invalid content of JSON fields")
		return http.StatusBadRequest, "invalid content of fields"
	case errors.Is(err, usernoterepo.ErrNotFound):
		log.Info().Msg("note was not found in repository")
		return http.StatusNotFound, "note not found"
	default:
		log.Error().Msg("service error")
		return http.StatusInternalServerError, "service error"
	}
}
