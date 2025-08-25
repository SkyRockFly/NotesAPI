package noterepository

import (
	"time"
)

type Note struct {
	ID        int        `db:"id"`
	AccountID int        `db:"account_id"`
	Title     string     `db:"title"`
	Body      string     `db:"body"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"` //time.Time
}
