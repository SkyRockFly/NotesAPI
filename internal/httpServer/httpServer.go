package httpserver

import (
	"NotesService/internal/middlewares"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	notes "NotesService/internal/notesRepository"

	"github.com/rs/zerolog"
)

var isShuttingDown atomic.Bool

func StartServer(ctx context.Context, port string, service *notes.RepositoryImpl) {
	onGoingCtx, cancelAll := context.WithCancel(ctx)
	mux := http.NewServeMux()
	mux.HandleFunc("/notes", middlewares.LogMiddleware(httpCreateNoteHandler(service)))
	mux.HandleFunc("/health", middlewares.LogMiddleware(checkHealthHandler()))

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
		BaseContext: func(_ net.Listener) context.Context {
			return onGoingCtx
		},
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("ListenAndServe error: %v", err)
		}
	}()

	<-ctx.Done()
	isShuttingDown.Store(true)
	onGoingCtx, cancel := context.WithTimeout(ctx, time.Second*7)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("shutdown error:%v", err)
	}
	cancelAll()
}

func httpCreateNoteHandler(service *notes.RepositoryImpl) middlewares.HandlerFuncWithStatus {
	return func(w http.ResponseWriter, r *http.Request) (middlewares.APIResponse, int, error) {
		if r.Method != http.MethodPost {
			return middlewares.APIResponse{Error: "method is now allowed"},
				http.StatusMethodNotAllowed, fmt.Errorf("httpServer.httpCreateNoteHandler:method is now allowed")
		}
		ctx := r.Context()

		var note notes.Note
		if err := json.NewDecoder(r.Body).Decode(&note); err != nil {
			return middlewares.APIResponse{Error: "invalid json"},
				http.StatusBadRequest, fmt.Errorf("httpCreateNoteHandler:%w", err)
		}

		if err := notes.JsonValidator(note); err != nil {
			return middlewares.APIResponse{Error: err.Error()},
				http.StatusBadRequest, fmt.Errorf("httpCreateNoteHandler:%w", err)
		}

		id, err := service.Create(ctx, note.AccountID, note.Title, note.Title)
		if err != nil {
			return middlewares.APIResponse{Error: "service error"},
				http.StatusInternalServerError, fmt.Errorf("httpCreateNoteHandler:%w", err)
		}

		resp := struct {
			ID int `json:"id"`
		}{ID: id}

		return middlewares.APIResponse{Data: resp}, http.StatusCreated, nil
	}
}

func checkHealthHandler() middlewares.HandlerFuncWithStatus {
	return func(w http.ResponseWriter, r *http.Request) (middlewares.APIResponse, int, error) {
		if isShuttingDown.Load() {
			return middlewares.APIResponse{Error: "shutting down"},
				http.StatusServiceUnavailable, fmt.Errorf("httpServer.checkHealthHandler:shutting down")
		}
		_, err := w.Write([]byte("ok"))
		if err != nil {
			logger := r.Context().Value(middlewares.LoggerCtxKey).(zerolog.Logger)
			logger.Warn().Err(err).Msg("failed to write response")
			return middlewares.APIResponse{}, 0, fmt.Errorf("checkHealthHandler:%w", err)
		}
		resp := struct {
			Health bool `json:"health"`
		}{Health: true}
		return middlewares.APIResponse{Data: resp}, http.StatusOK, nil
	}
}
