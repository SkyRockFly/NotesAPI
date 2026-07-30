package httpserver

import (
	"fmt"
	"net/http"
	notesvc "notes/internal/gateway/service/note"
	"notes/internal/pkg/middlewares"
)

type deleteDTO struct {
	ID int `json:"id"`
}

func deleteNoteHandler(service *notesvc.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := getCtxLogger(ctx)

		var dtoNote deleteDTO
		if err := decodeJSON(&dtoNote, r); err != nil {
			handleError(w, fmt.Errorf("decode: %w", err), logger)
			return
		}

		uid, ok := ctx.Value(middlewares.UIDKey).(int)
		if !ok {
			handleError(w, fmt.Errorf("cant extract uid value"), logger)
			return
		}

		note := notesvc.DeleteReq{
			ID:        dtoNote.ID,
			AccountID: uid,
		}
		if err := service.Delete(ctx, note); err != nil {
			handleError(w, fmt.Errorf("svc.delete: %w", err), logger)
			return
		}

		writeJSON(w, http.StatusNoContent, logger, nil)
	}
}
