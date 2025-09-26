package httpserver

import (
	"net/http"
	noteservice "notes/internal/notes-svc/noteService"
)

type DeleteNoteResp struct {
	Deleted bool `json:"deleted"`
}

func HTTPDeleteNoteHandler(service *noteservice.Service) http.HandlerFunc {
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
		if err := service.Delete(ctx, note); err != nil {
			code, info := mapToHTTPError(err, logger)
			logger.
				Warn().
				Err(err).
				Msg("service error")
			writeJSON(w, code,
				logger, info)
			return
		}

		writeJSON(w, http.StatusOK, logger, DeleteNoteResp{Deleted: true})
	}
}
