package httpserver

import (
	"fmt"
	"net/http"
	notesvc "notes/internal/gateway/service/note"
	"notes/internal/pkg/middlewares"
)

type getDTO struct {
	ID int `json:"id"`
}

func getNoteHandler(service *notesvc.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := getCtxLogger(ctx)

		var dtoNote getDTO
		if err := decodeJSON(&dtoNote, r); err != nil {
			handleError(w, fmt.Errorf("decode: %w", err), logger)
			return
		}

		uid, ok := ctx.Value(middlewares.UIDKey).(int)
		if !ok {
			handleError(w, fmt.Errorf("can't extract uid value"), logger)
			return
		}

		note := notesvc.GetReq{
			ID:        dtoNote.ID,
			AccountID: uid,
		}

		svcNote, err := service.Get(ctx, note)
		if err != nil {
			handleError(w, fmt.Errorf("svc.get: %w", err), logger)
			return
		}

		respNote := remapSvcToRespNote(svcNote)
		writeJSON(w, http.StatusOK, logger, respNote)
	}
}
