package httpserver

import (
	"fmt"
	"net/http"
	userservice "notes/internal/gateway/userService"
)

type getDTO struct {
	ID        int `json:"id"`
	AccountID int `json:"account_id"`
}

func getNoteHandler(service *userservice.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := getCtxLogger(ctx)

		dtoNote, err := decodeJSON[getDTO](r)
		if err != nil {
			logger.Info().Err(fmt.Errorf("decode: %w", err)).Msg("get handler")
			writeJSON(w, http.StatusBadRequest, logger, errorAPIResponse{Err: "invalid json"})
			return
		}

		note := userservice.GetNoteData{
			ID:        dtoNote.ID,
			AccountID: dtoNote.AccountID,
		}

		svcNote, err := service.Get(ctx, note)
		if err != nil {
			code, desc := mapToHTTPError(err, logger, "get handler")
			writeJSON(w, code, logger, errorAPIResponse{Err: desc})
			return
		}

		respNote := remapSvcToRespNote(svcNote)
		writeJSON(w, http.StatusOK, logger, respNote)
	}
}
