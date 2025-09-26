package httpserver

import (
	"net/http"
	noteservice "notes/internal/notes-svc/noteService"
)

type UpdateNoteResp struct {
	Updated bool `json:"updated"`
}

func HTTPUpdateNoteHandler(service *noteservice.Service) http.HandlerFunc {
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
		if err := service.Update(ctx, note); err != nil {
			code, info := mapToHTTPError(err, logger)
			logger.
				Warn().
				Err(err).
				Msg("service error")
			writeJSON(w, code,
				logger, info)
			return
		}

		writeJSON(w, http.StatusOK,
			logger, UpdateNoteResp{Updated: true})
	}
}
