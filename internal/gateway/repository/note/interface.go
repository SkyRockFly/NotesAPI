package noterepo

import (
	"context"
	"time"
)

type Note struct {
	ID        int       `json:"id"`
	AccountID int       `json:"account_id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListReq struct {
	AccountID int  `json:"account_id"`
	Limit     int  `json:"limit"`
	Cursor    int  `json:"cursor"`
	Next      bool `json:"next"`
}

type ListResp struct {
	CursorNext int    `json:"cursor_next"`
	CursorPrev int    `json:"cursor_prev"`
	Notes      []Note `json:"notes"`
	HasMore    bool   `json:"has_more"`
}

type GetReq struct {
	ID        int `json:"id"`
	AccountID int `json:"account_id"`
}

type CreateReq struct {
	AccountID int    `json:"account_id"`
	Title     string `json:"title"`
	Body      string `json:"body"`
}

type DeleteReq struct {
	ID        int `json:"id"`
	AccountID int `json:"account_id"`
}

type UpdateReq struct {
	ID        int    `json:"id"`
	AccountID int    `json:"account_id"`
	Title     string `json:"title"`
	Body      string `json:"body"`
}

type INote interface {
	Get(ctx context.Context, req GetReq) (Note, error)
	Create(ctx context.Context, req CreateReq) (int, error)
	Update(ctx context.Context, req UpdateReq) error
	List(ctx context.Context, req ListReq) (ListResp, error)
	Delete(ctx context.Context, req DeleteReq) error
}
