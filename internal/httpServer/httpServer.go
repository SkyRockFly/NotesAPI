package httpserver

import (
	"NotesService/internal/middlewares"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	noteservice "NotesService/internal/noteService"
	notetype "NotesService/internal/noteType"

	"github.com/rs/zerolog"
)

var isShuttingDown atomic.Bool

type errorAPIResponse struct {
	Err string `json:"err"`
}

type healthResponse struct {
	Health bool `json:"health"`
}

type CreateNoteResp struct {
	ID int `json:"id"`
}

type DeleteNoteResp struct {
	Deleted bool `json:"deleted"`
}

func StartServer(ctx context.Context, port string, service *noteservice.Service) {
	mux := http.NewServeMux()
	mux.HandleFunc("/notes", middlewares.LogMiddleware(httpCreateNoteHandler(service)))
	mux.HandleFunc("/health", middlewares.LogMiddleware(healthCheckHandler()))

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			panic(fmt.Sprintf("ListenAndServe error: %v", err))
		}
	}()

	<-ctx.Done()
	isShuttingDown.Store(true)

	shutdownCtx, done := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer done()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error:%v", err)
	}
}

func httpCreateNoteHandler(service *noteservice.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := r.Context().Value(middlewares.LoggerCtxKey).(zerolog.Logger)
		if r.Method != http.MethodPost {
			logger.Warn().Str("httpCreateNoteHandler", "method is not allowed")
			sendJSONError(errorAPIResponse{Err: "method is not allowed"}, w,
				http.StatusMethodNotAllowed)
			return
		}
		ctx := r.Context()

		var noteDTO notetype.DTO
		if err := json.NewDecoder(r.Body).Decode(&noteDTO); err != nil {
			logger.Warn().Str("httpCreateNoteHandler:", "invalid json")
			sendJSONError(errorAPIResponse{Err: "invalid json"}, w, http.StatusBadRequest)
			return
		}

		note := noteservice.RemapDTOtoServ(noteDTO)
		id, err := service.Create(ctx, note.AccountID, note.Title, note.Body)
		if err != nil {
			logger.Warn().Str("httpCreateNoteHandler", "service error")
			sendJSONError(errorAPIResponse{Err: "service error"}, w,
				http.StatusInternalServerError)
			return
		}

		resp := &CreateNoteResp{ID: id}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			logger.Warn().Str("httpCreateNoteHandler", "Encode failed")
		}
	}
}

func httpDeleteNoteHandler(service *noteservice.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := r.Context().Value(middlewares.LoggerCtxKey).(zerolog.Logger)
		if r.Method != http.MethodGet {
			logger.Warn().Str("httpDeleteNoteHandler", "method is not allowed")
			sendJSONError(errorAPIResponse{Err: "method is not allowed"}, w,
				http.StatusMethodNotAllowed)
			return
		}
		ctx := r.Context()
		var noteDTO notetype.DTO
		if err := json.NewDecoder(r.Body).Decode(&noteDTO); err != nil {
			logger.Warn().Str("httpDeleteNoteHandler:", "invalid json")
			sendJSONError(errorAPIResponse{Err: "invalid json"}, w, http.StatusBadRequest)
			return
		}

		note := noteservice.RemapDTOtoServ(noteDTO)
		err := service.Delete(ctx, note.ID, note.AccountID)
		if err != nil {
			logger.Warn().Str("httpCreateNoteHandler", "service error")
			sendJSONError(errorAPIResponse{Err: "service error"}, w,
				http.StatusInternalServerError)
			return
		}

		resp := &DeleteNoteResp{Deleted: true}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			logger.Warn().Str("httpCreateNoteHandler", "Encode failed")
		}

	}
}

func healthCheckHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := r.Context().Value(middlewares.LoggerCtxKey).(zerolog.Logger)
		if isShuttingDown.Load() {
			sendJSONError(errorAPIResponse{Err: "shutting down"}, w, http.StatusServiceUnavailable)
			logger.Warn().Msg("httpServer.healthcheckHandler: shutting down")
			return
		}
		resp := &healthResponse{Health: true}
		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			logger.Warn().Err(err).Msg("httpServer.healthcheckHandler: json.Encode failed")
			return
		}
	}
}

func sendJSONError(err errorAPIResponse, w http.ResponseWriter, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(err); err != nil {
		log.Printf("httpServer.sendJsonError: %v", err)
	}
}
