package httpserver

import (
	"NotesService/internal/middlewares"
	noterepository "NotesService/internal/noteRepository"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	noteservice "NotesService/internal/noteService"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	maxBodySize = 10 << 20
	encodeError = `{"error":"encode failed"}`
)

var isShuttingDown atomic.Bool

type errorAPIResponse struct {
	Err string `json:"error"`
}

type healthResponse struct {
	Health bool `json:"health"`
}

type DeleteNoteResp struct {
	Deleted bool `json:"deleted"`
}

type UpdateNoteResp struct {
	Updated bool `json:"updated"`
}

type DTO struct {
	ID        int    `json:"id,omitempty"`
	AccountID int    `json:"account_id,omitempty"`
	Title     string `json:"title,omitempty"`
	Body      string `json:"body,omitempty"`
}

type NoteResponse struct {
	ID        int       `json:"id"`
	AccountID int       `json:"account_id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func StartServer(ctx context.Context, port string, service *noteservice.Service) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /note/create", middlewares.LogMiddleware(HTTPCreateNoteHandler(service)))
	mux.HandleFunc("DELETE /note/delete", middlewares.LogMiddleware(HTTPDeleteNoteHandler(service)))
	mux.HandleFunc("PUT /note/update", middlewares.LogMiddleware(HTTPUpdateNoteHandler(service)))
	mux.HandleFunc("GET /note/get", middlewares.LogMiddleware(HTTPGetNoteHandler(service)))
	mux.HandleFunc("/health", middlewares.LogMiddleware(HealthCheckHandler()))
	mux.HandleFunc("GET /notes/get", middlewares.LogMiddleware(HTTPListNoteHandler(service)))

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Panic().Err(fmt.Errorf("listenAndServe: %w", err)).Msg("httpserver.start server") //return error
		}
	}()

	<-ctx.Done()
	isShuttingDown.Store(true)

	shutdownCtx, done := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer done()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Warn().Err(err).Msg("shutdown")
	}
}

func HTTPDeleteNoteHandler(service *noteservice.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "unsupported media type", http.StatusUnsupportedMediaType)
			return
		}
		ctx := r.Context()
		logger := getCtxLogger(ctx)

		noteDTO, err := parseJSON(w, r)
		if err != nil {
			logger.
				Warn().
				Str("parse", "invalid json")
			writeJSON(w, http.StatusBadRequest,
				logger, errorAPIResponse{Err: "invalid json"})
			return
		}

		note := remapDTOtoServ(noteDTO)
		if err := service.Delete(ctx, note); err != nil {
			code, info := mapHTTPError(err)
			logger.
				Warn().
				Err(err).
				Msg("service error")
			writeJSON(w, code,
				logger, info)
			return
		}

		writeJSON(w, http.StatusOK, logger, DeleteNoteResp{Deleted: true})
	}
}

func HTTPUpdateNoteHandler(service *noteservice.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "unsupported media type", http.StatusUnsupportedMediaType)
			return
		}
		ctx := r.Context()
		logger := getCtxLogger(ctx)

		noteDTO, err := parseJSON(w, r)
		if err != nil {
			logger.
				Warn().
				Str("parse", "invalid json")
			writeJSON(w, http.StatusBadRequest,
				logger, errorAPIResponse{Err: "invalid json"})
			return
		}

		note := remapDTOtoServ(noteDTO)
		if err := service.Update(ctx, note); err != nil {
			code, info := mapHTTPError(err)
			logger.
				Warn().
				Err(err).
				Msg("service error")
			writeJSON(w, code,
				logger, info)
			return
		}

		writeJSON(w, http.StatusOK,
			logger, UpdateNoteResp{Updated: true})
	}
}

func HTTPGetNoteHandler(service *noteservice.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "unsupported media type", http.StatusUnsupportedMediaType)
			return
		}
		ctx := r.Context()
		logger := getCtxLogger(ctx)

		noteDTO, err := parseJSON(w, r)
		if err != nil {
			logger.
				Warn().
				Str("parse", "invalid json")
			writeJSON(w, http.StatusBadRequest,
				logger, errorAPIResponse{Err: "invalid json"})
			return
		}
		note := remapDTOtoServ(noteDTO)

		servNote, err := service.Get(ctx, note)
		if err != nil {
			code, info := mapHTTPError(err)
			logger.
				Warn().
				Err(err).
				Msg("service error")
			writeJSON(w, code,
				logger, info)
			return
		}
		newNote := remapServToResp(servNote)

		writeJSON(w, http.StatusOK,
			logger, newNote)
	}
}

func HTTPListNoteHandler(service *noteservice.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "unsupported media type", http.StatusUnsupportedMediaType)
			return
		}
		ctx := r.Context()
		logger := getCtxLogger(ctx)

		noteDTO, err := parseJSON(w, r)
		if err != nil {
			logger.
				Warn().
				Str("parse", "invalid json")
			writeJSON(w, http.StatusBadRequest,
				logger, errorAPIResponse{Err: "invalid json"})
			return
		}
		note := remapDTOtoServ(noteDTO)

		notes, err := service.List(ctx, note)
		if err != nil {
			code, info := mapHTTPError(err)
			logger.
				Warn().
				Err(err).
				Msg("service error")
			writeJSON(w, code,
				logger, info)
			return
		}
		respNotes := remapListToResp(notes)
		writeJSON(w, http.StatusOK,
			logger, respNotes)
	}
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

func parseJSON(w http.ResponseWriter, r *http.Request) (DTO, error) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	var note DTO
	if err := dec.Decode(&note); err != nil {
		return DTO{}, fmt.Errorf("parse: %w", err)
	}
	return note, nil
}

func writeJSON(w http.ResponseWriter, statusCode int, logger *zerolog.Logger, resp any) {
	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(resp); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(encodeError))
		logger.Warn().Err(err).Msg("encode")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if _, err := w.Write(buf.Bytes()); err != nil {
		logger.Warn().Err(err).Msg("write")
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

func remapDTOtoServ(noteDTO DTO) *noteservice.Note {
	serviceNote := &noteservice.Note{
		ID:        noteDTO.ID,
		AccountID: noteDTO.AccountID,
		Title:     noteDTO.Title,
		Body:      noteDTO.Body,
	}
	return serviceNote
}

func remapServToResp(servNote *noteservice.Note) *NoteResponse {
	respNote := &NoteResponse{
		ID:        servNote.ID,
		AccountID: servNote.AccountID,
		Title:     servNote.Title,
		Body:      servNote.Body,
		CreatedAt: servNote.CreatedAt,
		UpdatedAt: servNote.UpdatedAt,
	}
	return respNote
}

func mapHTTPError(err error) (int, any) {
	switch {
	case errors.Is(err, noteservice.ErrInvalid):
		return http.StatusBadRequest, errorAPIResponse{Err: "invalid json"}
	case errors.Is(err, noterepository.ErrNotFound):
		return http.StatusNotFound, errorAPIResponse{Err: "note not found"}
	default:
		return http.StatusInternalServerError, errorAPIResponse{Err: "service error"}
	}
}
