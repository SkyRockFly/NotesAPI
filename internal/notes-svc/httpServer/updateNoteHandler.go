package httpserver

import (
	"fmt"
	"net/http"
	noteservice "notes/internal/notes-svc/noteService"
)

type UpdateDTO struct {
	ID        int    `json:"id"`
	AccountID int    `json:"account_id"`
	Title     string `json:"title"`
	Body      string `json:"body"`
}

func HTTPUpdateNoteHandler(service *noteservice.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := getCtxLogger(ctx)

		var dto UpdateDTO
		if err := decodeJSON(&dto, r); err != nil {
			handleError(w, fmt.Errorf("decode:%w", err), logger)
			return
		}

		updateReq := noteservice.UpdateReq{
			AccountID: dto.AccountID,
			ID:        dto.ID,
			Title:     dto.Title,
			Body:      dto.Body,
		}

		if err := service.Update(ctx, updateReq); err != nil {
			handleError(w, fmt.Errorf("svc.Update:%w", err), logger)
			return
		}

		writeJSON(w, http.StatusNoContent,
			logger, nil)
	}
}
