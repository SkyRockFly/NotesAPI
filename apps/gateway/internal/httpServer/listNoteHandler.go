package httpserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	userservice "notes/apps/gateway/internal/userService"
)

type listDTO struct {
	AccountID int `json:"account_id"`
}

func listNoteHandler(service *userservice.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := getCtxLogger(ctx)

		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		var dtoNote listDTO
		if err := dec.Decode(&dtoNote); err != nil {
			logger.Info().Err(fmt.Errorf("decode: %w", err)).Msg("list handler")
			writeJSON(w, http.StatusBadRequest, logger, errorAPIResponse{Err: "invalid json"})
			return
		}

		note := userservice.Note{
			AccountID: dtoNote.AccountID,
		}

		svcNotes, err := service.Get(ctx, note) //change
		if err != nil {
			logger.Info().Err(fmt.Errorf("service: %w", err)).Msg("list handler")
			code, desc := mapToHTTPError(err, logger)
			writeJSON(w, code, logger, errorAPIResponse{Err: desc})
		}

		respNote := remapSvcToRespNote(svcNotes)
		writeJSON(w, http.StatusOK, logger, respNote)
	}
}
