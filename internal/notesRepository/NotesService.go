package notes

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Note struct {
	Id        int        `json:"id" db:"id"`
	AccountId int        `json:"account_id" db:"account_id"`
	Title     string     `json:"title" db:"title"`
	Body      string     `json:"body" db:"body"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

type UpdateNote struct {
	id    int     `json:"id" db:"id"`
	title *string `json:"title" db:"title"`
	body  *string `json:"body" db:"body"`
}

type NotesRepository interface {
	Create(ctx context.Context, account_id int, title string, body string) (int, error)
	Delete(id int) error
	Get(id int) (Note, error)
	List(account_id int) ([]Note, error)
	Update(newNote UpdateNote) (int, error)
}

type NotesRepositoryImpl struct {
	repo NotesRepository
}

func NotesNewRepositoryImpl(repo NotesRepository) *NotesRepositoryImpl {
	return &NotesRepositoryImpl{repo: repo}
}

func (s *NotesRepositoryImpl) Create(ctx context.Context, account_id int, title string, body string) (int, error) {
	return s.repo.Create(ctx, account_id, title, body)
}

func (s *NotesRepositoryImpl) Delete(id int) error {
	return s.repo.Delete(id)
}

func (s *NotesRepositoryImpl) Get(id int) (Note, error) {
	return s.repo.Get(id)
}

func (s *NotesRepositoryImpl) List(account_id int) ([]Note, error) {
	return s.repo.List(account_id)
}

func (s *NotesRepositoryImpl) Update(newNote UpdateNote) (int, error) {
	return s.repo.Update(newNote)
}

type PostgresNotesImpl struct{ postgresDB *pgxpool.Pool }

func PostgresNewRepository(postgresDB *pgxpool.Pool) *PostgresNotesImpl {
	return &PostgresNotesImpl{postgresDB: postgresDB}
}

func (s *PostgresNotesImpl) Create(ctx context.Context, account_id int, title string, body string) (int, error) {
	sql := `INSERT INTO note (account_id, title, body)
			VALUES ($1, $2, $3)
			RETURNING id`
	row := s.postgresDB.QueryRow(ctx, sql, account_id, title, body)
	var id int
	err := row.Scan(&id)
	if err != nil {
		err = fmt.Errorf("func notes.Create(): failed to find id %w", err)
	}

	return id, err
}

func (s *PostgresNotesImpl) Delete(id int) error {
	panic("delete not implemented yet")
}

func (s *PostgresNotesImpl) Get(id int) (Note, error) {
	panic("get not implemented yet")
}

func (s *PostgresNotesImpl) List(account_id int) ([]Note, error) {
	panic("list not implemented yet")
}

func (s *PostgresNotesImpl) Update(newNote UpdateNote) (int, error) {
	panic("update not implemented yet")
}
