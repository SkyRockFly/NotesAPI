package httpserver

import (
	"fmt"
	"net/http"
	noteservice "notes/internal/notes-svc/noteService"
)

type DeleteDTO struct {
	ID        int `json:"id"`
	AccountID int `json:"account_id"`
}

func HTTPDeleteNoteHandler(service *noteservice.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := getCtxLogger(ctx)

		var dto DeleteDTO
		if err := decodeJSON(&dto, r); err != nil {
			handleError(w, fmt.Errorf("decode:%w", err), logger)
			return
		}

		deleteReq := noteservice.DeleteReq{
			ID:        dto.ID,
			AccountID: dto.AccountID,
		}
		if err := service.Delete(ctx, deleteReq); err != nil {
			handleError(w, fmt.Errorf("svc.Delete:%w", err), logger)
			return
		}

		writeJSON(w, http.StatusNoContent, logger, nil)
	}
}
