package httpserver

import (
	"net/http"
	noteservice "notes/internal/notes-svc/noteService"
)

func HTTPListNoteHandler(service *noteservice.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := getCtxLogger(ctx)

		noteDTO, err := parseJSON(r)
		if err != nil {
			logger.
				Warn().
				Str("parse", "invalid json")
			writeJSON(w, http.StatusBadRequest,
				logger, errorAPIResponse{Err: "invalid json"})
			return
		}
		note := remapDTOtoSVC(noteDTO)

		notes, err := service.List(ctx, note)
		if err != nil {
			code, info := mapToHTTPError(err, logger)
			logger.
				Warn().
				Err(err).
				Msg("service error")
			writeJSON(w, code,
				logger, info)
			return
		}
		respNotes := remapListToResp(notes)
		writeJSON(w, http.StatusOK,
			logger, respNotes)
	}
}
