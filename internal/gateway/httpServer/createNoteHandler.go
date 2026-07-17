package httpserver

import (
	"fmt"
	"net/http"
	notesvc "notes/internal/gateway/service/note"
	"notes/internal/pkg/middlewares"
)

type createDTO struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type createResponse struct {
	ID int64 `json:"id"`
}

func createNoteHandler(service *notesvc.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := getCtxLogger(ctx)

		var dtoNote createDTO
		if err := decodeJSON(&dtoNote, r); err != nil {
			handleError(w, fmt.Errorf("decode:%w", err), logger)
			return
		}

		uid, ok := ctx.Value(middlewares.UIDKey).(int)
		if !ok {
			handleError(w, fmt.Errorf("cant extract ctx value"), logger)
			return
		}

		note := notesvc.CreateReq{
			AccountID: uid,
			Title:     dtoNote.Title,
			Body:      dtoNote.Body,
		}

		id, err := service.Create(ctx, note)
		if err != nil {
			handleError(w, fmt.Errorf("svc.create: %w", err), logger)
			return
		}

		writeJSON(w, http.StatusCreated, logger, createResponse{ID: id})
	}
}
