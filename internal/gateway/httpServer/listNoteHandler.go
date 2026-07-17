package httpserver

import (
	"fmt"
	"net/http"
	notesvc "notes/internal/gateway/service/note"
	"notes/internal/pkg/middlewares"
)

type ListDTO struct {
	Cursor int64 `json:"cursor"`
	Limit  int   `json:"limit"`
	Next   bool  `json:"next"`
}

type ListResp struct {
	CursorNext int64          `json:"cursor_next"`
	CursorPrev int64          `json:"cursor_prev"`
	Notes      []ResponseNote `json:"notes"`
	HasMore    bool           `json:"has_more"`
}

func listNoteHandler(service *notesvc.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := getCtxLogger(ctx)

		var dtoNote ListDTO
		if err := decodeJSON(&dtoNote, r); err != nil {
			handleError(w, fmt.Errorf("decode: %w", err), logger)
			return
		}

		uid, ok := ctx.Value(middlewares.UIDKey).(int)
		if !ok {
			handleError(w, fmt.Errorf("can't extract uid value"), logger)
			return
		}

		note := notesvc.ListReq{
			AccountID: uid,
			Cursor:    dtoNote.Cursor,
			Limit:     dtoNote.Limit,
			Next:      dtoNote.Next,
		}

		svcResp, err := service.List(ctx, note)
		if err != nil {
			handleError(w, fmt.Errorf("svc.list: %w", err), logger)
			return
		}

		respNotes := make([]ResponseNote, 0, len(svcResp.Notes))
		var respNote ResponseNote
		for _, svcNote := range svcResp.Notes {
			respNote = remapSvcToRespNote(svcNote)
			respNotes = append(respNotes, respNote)
		}

		resp := ListResp{
			CursorNext: svcResp.CursorNext,
			CursorPrev: svcResp.CursorPrev,
			HasMore:    svcResp.HasMore,
			Notes:      respNotes,
		}

		writeJSON(w, http.StatusOK, logger, resp)
	}
}
