package httpserver

import (
	"fmt"
	"net/http"
	notesvc "notes/internal/gateway/service/note"
	"notes/internal/pkg/middlewares"
)

type updateDTO struct {
	ID    int    `json:"ID"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

type updateResponse struct {
	Updated bool `json:"updated"`
}

func updateNoteHandler(service *notesvc.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := getCtxLogger(ctx)

		var dtoNote updateDTO
		if err := decodeJSON(&dtoNote, r); err != nil {
			handleError(w, fmt.Errorf("decode: %w", err), logger)
			return
		}

		uid, ok := ctx.Value(middlewares.UIDKey).(int)
		if !ok {
			handleError(w, fmt.Errorf("can't extract UIDKey"), logger)
			return
		}

		note := notesvc.UpdateReq{
			ID:        dtoNote.ID,
			AccountID: uid,
			Title:     dtoNote.Title,
			Body:      dtoNote.Body,
		}

		if err := service.Update(ctx, note); err != nil {
			handleError(w, fmt.Errorf("svc.update: %w", err), logger)
			return
		}

		writeJSON(w, http.StatusNoContent, logger, nil)
	}
}
