package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"NotesService/internal/middlewares"
	notes "NotesService/internal/notesRepository"
)

var isShuttingDown atomic.Bool

func StartServer(ctx context.Context, port string, service *notes.NotesRepositoryImpl) {
	onGoingCtx, cancelAll := context.WithCancel(ctx)
	mux := http.NewServeMux()
	mux.HandleFunc("/notes", middlewares.LogMiddleware(httpCreateNoteHandler(service)))
	mux.HandleFunc("/health", checkHealthHandler)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
		BaseContext: func(l net.Listener) context.Context {
			return onGoingCtx
		},
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()

	<-ctx.Done()
	isShuttingDown.Store(true)
	onGoingCtx, cancel := context.WithTimeout(ctx, time.Second*7)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown error:%v", err)
	}
	cancelAll()

}

func httpCreateNoteHandler(service *notes.NotesRepositoryImpl) middlewares.HandlerFuncWithStatus {
	return func(w http.ResponseWriter, r *http.Request) (middlewares.APIResponse, int, error) {
		if r.Method != http.MethodPost {
			return middlewares.APIResponse{Error: "method is now allowed"},
				http.StatusMethodNotAllowed, errors.New("method is now allowed")
		}
		ctx := r.Context()

		var note notes.Note
		if err := json.NewDecoder(r.Body).Decode(&note); err != nil {
			return middlewares.APIResponse{Error: "invalid json"},
				http.StatusBadRequest, errors.New("invalid json")
		}

		id, err := service.Create(ctx, note.AccountId, note.Title, note.Title)
		if err != nil {
			return middlewares.APIResponse{Error: "service error"},
				http.StatusInternalServerError, errors.New("service error")
		}

		resp := struct {
			ID int `json:"id"`
		}{ID: id}

		return middlewares.APIResponse{Data: resp}, http.StatusCreated, nil
	}

}

func checkHealthHandler(w http.ResponseWriter, r *http.Request) {
	if isShuttingDown.Load() {
		http.Error(w, "Shutting down", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
