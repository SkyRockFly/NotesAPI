package httpserver

import (
	noteservice "NotesService/internal/noteService"
	"net/http"
)

func HTTPGetNoteHandler(service *noteservice.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "unsupported media type", http.StatusUnsupportedMediaType)
			return
		}
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

		servNote, err := service.Get(ctx, note)
		if err != nil {
			code, info := mapHTTPError(err)
			logger.
				Warn().
				Err(err).
				Msg("service error")
			writeJSON(w, code,
				logger, info)
			return
		}
		newNote := remapSVCToResp(servNote)

		writeJSON(w, http.StatusOK,
			logger, newNote)
	}
}
