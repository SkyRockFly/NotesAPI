package httpserver

import (
	"fmt"
	"net/http"
	noteservice "notes/internal/notes-svc/noteService"
)

type CreateNoteResp struct {
	ID int `json:"id"`
}

type CreateDTO struct {
	AccountID int    `json:"account_id"`
	Title     string `json:"title"`
	Body      string `json:"body"`
}

func HTTPCreateNoteHandler(service *noteservice.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := getCtxLogger(ctx)

		var dto CreateDTO
		if err := decodeJSON(&dto, r); err != nil {
			handleError(w, fmt.Errorf("decode:%w", err), logger)
			return
		}

		createReq := noteservice.CreateReq{
			AccountID: dto.AccountID,
			Title:     dto.Title,
			Body:      dto.Body,
		}
		id, err := service.Create(ctx, createReq)
		if err != nil {
			handleError(w, fmt.Errorf("svc.Create:%w", err), logger)
			return
		}
		writeJSON(w, http.StatusCreated,
			logger, CreateNoteResp{ID: id})
	}
}
