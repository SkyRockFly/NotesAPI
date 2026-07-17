package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"notes/internal/pkg/apperror"
	"notes/internal/pkg/middlewares"
	"sync/atomic"
	"time"

	noteservice "notes/internal/notes-svc/noteService"

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

type NoteResponse struct {
	ID        int64     `json:"id"`
	AccountID int       `json:"account_id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func StartServer(ctx context.Context, port string, service *noteservice.Service) error {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /note/create",
		middlewares.LogMiddleware(
			middlewares.JSONReqSizeMiddleware(
				middlewares.DemandJSONHeaders(
					HTTPCreateNoteHandler(service)))))

	mux.HandleFunc("DELETE /note/delete",
		middlewares.LogMiddleware(
			middlewares.JSONReqSizeMiddleware(
				middlewares.DemandJSONHeaders(
					HTTPDeleteNoteHandler(service)))))

	mux.HandleFunc("PUT /note/update",
		middlewares.LogMiddleware(
			middlewares.JSONReqSizeMiddleware(
				middlewares.DemandJSONHeaders(
					HTTPUpdateNoteHandler(service)))))

	mux.HandleFunc("POST /note/get",
		middlewares.LogMiddleware(
			middlewares.JSONReqSizeMiddleware(
				middlewares.DemandJSONHeaders(
					HTTPGetNoteHandler(service)))))

	mux.HandleFunc("POST /notes/get",
		middlewares.LogMiddleware(
			middlewares.JSONReqSizeMiddleware(
				middlewares.DemandJSONHeaders(
					HTTPListNoteHandler(service)))))

	mux.HandleFunc("/health",
		middlewares.LogMiddleware(
			HealthCheckHandler()))

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

func remapListToResp(notes []noteservice.Note) []NoteResponse {
	servNotes := make([]NoteResponse, 0, len(notes))

	for _, note := range notes {
		respNote := &NoteResponse{
			ID:        note.ID,
			AccountID: note.AccountID,
			Title:     note.Title,
			Body:      note.Body,
			CreatedAt: note.CreatedAt,
			UpdatedAt: note.UpdatedAt,
		}
		servNotes = append(servNotes, *respNote)
	}
	return servNotes
}

func decodeJSON(str any, r *http.Request) error {
	if str == nil {
		return fmt.Errorf("decode: empty struct")
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(str); err != nil {
		return fmt.Errorf("%w, decode: %w", apperror.ErrInvalidJSON, err)
	}

	return nil
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
	if statusCode == http.StatusNoContent {
		return
	}
	if _, err := w.Write(buf.Bytes()); err != nil {
		logger.Error().Err(fmt.Errorf("write: %w", err)).Msg("write JSON")
		return
	}
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

func remapSVCToResp(servNote noteservice.Note) NoteResponse {
	respNote := NoteResponse{
		ID:        servNote.ID,
		AccountID: servNote.AccountID,
		Title:     servNote.Title,
		Body:      servNote.Body,
		CreatedAt: servNote.CreatedAt,
		UpdatedAt: servNote.UpdatedAt,
	}
	return respNote
}

func handleError(w http.ResponseWriter, err error, log *zerolog.Logger) {
	var (
		code  int
		resp  errorAPIResponse
		level = zerolog.InfoLevel
	)
	switch {
	case errors.Is(err, apperror.ErrInvalidJSON):
		code = http.StatusUnprocessableEntity
		resp.Err = "invalid json"

	case errors.Is(err, apperror.ErrBadRequest):
		code = http.StatusBadRequest
		resp.Err = "bad request"

	case errors.Is(err, apperror.ErrNotFound):
		code = http.StatusNotFound
		resp.Err = "not found"

	case errors.Is(err, apperror.ErrBackend):
		level = zerolog.ErrorLevel
		code = http.StatusBadGateway
		resp.Err = "gateway error"

	case errors.Is(err, apperror.ErrAlreadyExists):
		code = http.StatusConflict
		resp.Err = "already exists"

	case errors.Is(err, apperror.ErrUnauthorized):
		code = http.StatusUnauthorized
		resp.Err = "unauthorized"
	default:
		level = zerolog.ErrorLevel
		code = http.StatusInternalServerError
		resp.Err = "service error"
	}

	log.WithLevel(level).
		Err(fmt.Errorf("response: %w", err)).
		Msg("request failed")

	writeJSON(w, code, log, resp)
}
