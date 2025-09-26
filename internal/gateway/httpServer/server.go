package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	usernoterepo "notes/internal/gateway/userNoteRepo"
	userservice "notes/internal/gateway/userService"
	"notes/internal/pkg/middlewares"
	"strconv"
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

func StartServer(ctx context.Context, port int, service *userservice.Service) error {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /note/get",
		middlewares.LogMiddleware(
			middlewares.JSONFileSizeMiddleware(
				middlewares.DemandJSONHeaders(
					getNoteHandler(service)))))

	mux.HandleFunc("DELETE /note/delete",
		middlewares.LogMiddleware(
			middlewares.JSONFileSizeMiddleware(
				middlewares.DemandJSONHeaders(
					deleteNoteHandler(service)))))

	mux.HandleFunc("POST /notes",
		middlewares.LogMiddleware(
			middlewares.JSONFileSizeMiddleware(
				middlewares.DemandJSONHeaders(
					listNoteHandler(service)))))

	mux.HandleFunc("PUT /note/update",
		middlewares.LogMiddleware(
			middlewares.JSONFileSizeMiddleware(
				middlewares.DemandJSONHeaders(
					updateNoteHandler(service)))))

	mux.HandleFunc("POST /note/create",
		middlewares.LogMiddleware(
			middlewares.JSONFileSizeMiddleware(
				middlewares.DemandJSONHeaders(
					createNoteHandler(service)))))

	server := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
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

func decodeJSON[T any](r *http.Request) (T, error) {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	var obj T
	if err := dec.Decode(&obj); err != nil {
		return obj, fmt.Errorf("decode: %w", err)
	}

	return obj, nil
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

func mapToHTTPError(err error, log *zerolog.Logger, hndName string) (int, string) {
	switch {
	case errors.Is(err, userservice.ErrInvalid):
		log.Info().Err(fmt.Errorf("service: %w", err)).Msg(hndName)
		return http.StatusBadRequest, "invalid content of fields"
	case errors.Is(err, usernoterepo.ErrNotFound):
		log.Info().Err(fmt.Errorf("service: %w", err)).Msg(hndName)
		return http.StatusNotFound, "note not found"
	default:
		log.Error().Err(fmt.Errorf("service: %w", err)).Msg(hndName)
		return http.StatusInternalServerError, "service error"
	}
}
