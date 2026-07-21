package httpserver

import (
	"fmt"
	"net/http"
	noteservice "notes/internal/notes-svc/noteService"
)

type ListDTO struct {
	AccountID int  `json:"account_id"`
	Limit     int  `json:"limit"`
	Cursor    int  `json:"cursor"`
	Next      bool `json:"next"`
}

type ListResp struct {
	CursorNext int            `json:"cursor_next"`
	CursorPrev int            `json:"cursor_prev"`
	Notes      []NoteResponse `json:"notes"`
	HasMore    bool           `json:"has_more"`
}

func HTTPListNoteHandler(service *noteservice.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := getCtxLogger(ctx)

		var dto ListDTO
		if err := decodeJSON(&dto, r); err != nil {
			handleError(w, fmt.Errorf("decode:%w", err), logger)
			return
		}

		listReq := noteservice.ListReq{
			AccountID: dto.AccountID,
			Limit:     dto.Limit,
			Cursor:    dto.Cursor,
			Next:      dto.Next,
		}

		svcResp, err := service.List(ctx, listReq)
		if err != nil {
			handleError(w, fmt.Errorf("svc.List:%w", err), logger)
			return
		}
		respNotes := remapListToResp(svcResp.Notes)

		resp := ListResp{
			CursorNext: svcResp.CursorNext,
			CursorPrev: svcResp.CursorPrev,
			Notes:      respNotes,
			HasMore:    svcResp.HasMore,
		}
		writeJSON(w, http.StatusOK,
			logger, resp)
	}
}
