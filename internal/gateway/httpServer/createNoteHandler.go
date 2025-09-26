package httpserver

import (
	"fmt"
	"net/http"
	userservice "notes/internal/gateway/userService"
)

type createDTO struct {
	AccountID int    `json:"account_id"`
	Title     string `json:"title"`
	Body      string `json:"body"`
}

type createResponse struct {
	ID int `json:"id"`
}

func createNoteHandler(service *userservice.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()
		logger := getCtxLogger(ctx)

		dtoNote, err := decodeJSON[createDTO](r)
		if err != nil {
			logger.Info().Err(fmt.Errorf("decode: %w", err)).Msg("create handler")
			writeJSON(w, http.StatusBadRequest, logger, errorAPIResponse{Err: "invalid json"})
			return
		}

		note := userservice.CreateNoteData{
			AccountID: dtoNote.AccountID,
			Title:     dtoNote.Title,
			Body:      dtoNote.Body,
		}

		id, err := service.Create(ctx, note)
		if err != nil {
			code, desc := mapToHTTPError(err, logger, "create handler")
			writeJSON(w, code, logger, errorAPIResponse{Err: desc})
			return
		}

		writeJSON(w, http.StatusCreated, logger, createResponse{ID: id})
	}
}
