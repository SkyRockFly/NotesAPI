package notetype

import "time"

type Repo struct {
	ID        int        `db:"id"`
	AccountID int        `db:"account_id"`
	Title     string     `db:"title"`
	Body      string     `db:"body"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

type DTO struct {
	ID        int    `json:"id"`
	AccountID int    `json:"account_id"`
	Title     string `json:"title"`
	Body      string `json:"body"`
}

type Update struct {
	id    int     ` db:"id"`
	title *string ` db:"title"`
	body  *string `db:"body"`
}

type Service struct {
	ID        int
	AccountID int
	Title     string
	Body      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
