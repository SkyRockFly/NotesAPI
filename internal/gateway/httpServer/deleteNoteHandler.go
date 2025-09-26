package httpserver

import (
	"fmt"
	"net/http"
	userservice "notes/internal/gateway/userService"
)

type deleteDTO struct {
	ID        int `json:"id"`
	AccountID int `json:"account_id"`
}

type deleteResponse struct {
	Deleted bool `json:"deleted"`
}

func deleteNoteHandler(service *userservice.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := getCtxLogger(ctx)

		dtoNote, err := decodeJSON[deleteDTO](r)
		if err != nil {
			logger.Info().Err(fmt.Errorf("decode: %w", err)).Msg("delete handler")
			writeJSON(w, http.StatusBadRequest, logger, errorAPIResponse{Err: "invalid json"})
			return
		}

		note := userservice.DeleteNoteData{
			ID:        dtoNote.ID,
			AccountID: dtoNote.AccountID,
		}

		deleted, err := service.Delete(ctx, note)
		if err != nil {
			code, desc := mapToHTTPError(err, logger, "delete handler")
			writeJSON(w, code, logger, errorAPIResponse{Err: desc})
			return
		}

		writeJSON(w, http.StatusOK, logger, deleteResponse{Deleted: deleted})
	}
}
