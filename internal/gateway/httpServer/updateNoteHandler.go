package httpserver

import (
	"fmt"
	"net/http"
	userservice "notes/internal/gateway/userService"
)

type updateDTO struct {
	ID        int    `json:"ID"`
	AccountID int    `json:"account_id"`
	Title     string `json:"title"`
	Body      string `json:"body"`
}

type updateResponse struct {
	Updated bool `json:"updated"`
}

func updateNoteHandler(service *userservice.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := getCtxLogger(ctx)

		dtoNote, err := decodeJSON[updateDTO](r)
		if err != nil {
			logger.Info().Err(fmt.Errorf("decode: %w", err)).Msg("update handler")
			writeJSON(w, http.StatusBadRequest, logger, errorAPIResponse{Err: "invalid json"})
			return
		}

		note := userservice.UpdateNoteData{
			ID:        dtoNote.ID,
			AccountID: dtoNote.AccountID,
			Title:     dtoNote.Title,
			Body:      dtoNote.Body,
		}

		updated, err := service.Update(ctx, note)
		if err != nil {
			code, desc := mapToHTTPError(err, logger, "update handler")
			writeJSON(w, code, logger, errorAPIResponse{Err: desc})
			return
		}

		writeJSON(w, http.StatusOK, logger, updateResponse{Updated: updated})
	}
}
