package httpserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	userservice "notes/apps/gateway/internal/userService"
)

type getDTO struct {
	ID        int `json:"id"`
	AccountID int `json:"account_id"`
}

func getNoteHandler(service *userservice.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := getCtxLogger(ctx)

		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		var dtoNote getDTO
		if err := dec.Decode(&dtoNote); err != nil {
			logger.Info().Err(fmt.Errorf("decode: %w", err)).Msg("get handler")
			writeJSON(w, http.StatusBadRequest, logger, errorAPIResponse{Err: "invalid json"})
			return
		}

		note := userservice.Note{
			ID:        dtoNote.ID,
			AccountID: dtoNote.AccountID,
		}

		svcNote, err := service.Get(ctx, note)
		if err != nil {
			logger.Info().Err(fmt.Errorf("service: %w", err)).Msg("get handler")
			code, desc := mapToHTTPError(err, logger)
			writeJSON(w, code, logger, errorAPIResponse{Err: desc})
		}

		respNote := remapSvcToRespNote(svcNote)
		writeJSON(w, http.StatusOK, logger, respNote)
	}
}
