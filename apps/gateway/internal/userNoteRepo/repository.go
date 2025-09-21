package usernoterepo

import (
	"errors"
	"time"
)

var (
	ErrNotFound   = errors.New("not found")   // 404
	ErrBadRequest = errors.New("bad request") // 400
	ErrUpstream   = errors.New("upstream")    // 500
)

type UserNote struct {
	ID        int
	AccountID int
	Title     string
	Body      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
