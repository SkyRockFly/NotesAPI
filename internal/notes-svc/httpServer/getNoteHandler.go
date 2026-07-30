package httpserver

import (
	"fmt"
	"net/http"
	noteservice "notes/internal/notes-svc/noteService"
)

type GetDTO struct {
	ID        int64 `json:"id"`
	AccountID int   `json:"account_id"`
}

func HTTPGetNoteHandler(service *noteservice.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := getCtxLogger(ctx)

		var dto DeleteDTO
		if err := decodeJSON(&dto, r); err != nil {
			handleError(w, fmt.Errorf("decode:%w", err), logger)
			return
		}

		getReq := noteservice.GetReq{
			ID:        dto.ID,
			AccountID: dto.AccountID,
		}

		servNote, err := service.Get(ctx, getReq)
		if err != nil {
			handleError(w, fmt.Errorf("svc.Get:%w", err), logger)
			return
		}
		newNote := remapSVCToResp(servNote)

		writeJSON(w, http.StatusOK,
			logger, newNote)
	}
}
