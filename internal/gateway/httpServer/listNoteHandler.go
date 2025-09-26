package httpserver

import (
	"fmt"
	"net/http"
	userservice "notes/internal/gateway/userService"
)

type listDTO struct {
	AccountID int `json:"account_id"`
}

func listNoteHandler(service *userservice.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := getCtxLogger(ctx)

		dtoNote, err := decodeJSON[listDTO](r)
		if err != nil {
			logger.Info().Err(fmt.Errorf("decode: %w", err)).Msg("list handler")
			writeJSON(w, http.StatusBadRequest, logger, errorAPIResponse{Err: "invalid json"})
			return
		}

		note := userservice.ListNoteData{
			AccountID: dtoNote.AccountID,
		}

		svcNotes, err := service.List(ctx, note)
		if err != nil {
			code, desc := mapToHTTPError(err, logger, "list handler")
			writeJSON(w, code, logger, errorAPIResponse{Err: desc})
			return
		}

		respNotes := make([]ResponseNote, 0, len(svcNotes))
		var respNote ResponseNote
		for _, svcNote := range svcNotes {
			respNote = remapSvcToRespNote(svcNote)
			respNotes = append(respNotes, respNote)
		}

		writeJSON(w, http.StatusOK, logger, respNotes)
	}
}
