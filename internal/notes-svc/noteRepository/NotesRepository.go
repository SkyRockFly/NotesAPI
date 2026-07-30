package noterepository

import (
	"context"
	"time"
)

type Note struct {
	ID        int       `db:"id"`
	AccountID int       `db:"account_id"`
	Title     string    `db:"title"`
	Body      string    `db:"body"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	DeletedAt time.Time `db:"deleted_at"`
}

type ListReq struct {
	AccountID int
	Limit     int
	Cursor    int
	Next      bool
}

type ListResp struct {
	CursorNext int
	CursorPrev int
	Notes      []Note
	HasMore    bool
}

type GetReq struct {
	ID        int
	AccountID int
}

type CreateReq struct {
	AccountID int
	Title     string
	Body      string
}

type DeleteReq struct {
	ID        int
	AccountID int
}

type UpdateReq struct {
	ID        int
	AccountID int
	Title     string
	Body      string
}

type IRepository interface {
	Create(ctx context.Context, req CreateReq) (int, error)
	Delete(ctx context.Context, req DeleteReq) error
	Get(ctx context.Context, req GetReq) (Note, error)
	List(ctx context.Context, req ListReq) (ListResp, error)
	Update(ctx context.Context, req UpdateReq) error
}
