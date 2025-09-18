package usernoterepo

import (
	"time"
)

type UserNote struct {
	ID        int
	AccountID int
	Title     string
	Body      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
