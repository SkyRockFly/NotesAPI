package notes

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Note struct {
	ID        int        `json:"id" db:"id"`
	AccountID int        `json:"account_id" db:"account_id"`
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

type Repository interface {
	Create(ctx context.Context, accountID int, title string, body string) (int, error)
	Delete(id int) error
	Get(id int) (Note, error)
	List(accountID int) ([]Note, error)
	Update(newNote UpdateNote) (int, error)
}

type RepositoryImpl struct {
	repo Repository
}

func NewRepositoryImpl(repo Repository) *RepositoryImpl {
	return &RepositoryImpl{repo: repo}
}

func (s *RepositoryImpl) Create(ctx context.Context, accountID int, title string, body string) (int, error) {
	id, err := s.repo.Create(ctx, accountID, title, body)
	if err != nil {
		return -1, fmt.Errorf("notesService.Create: %w", err)
	}
	return id, nil
}

func (s *RepositoryImpl) Delete(id int) error {
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("notesService.Delete:%w", err)
	}
	return nil
}

func (s *RepositoryImpl) Get(id int) (Note, error) {
	note, err := s.repo.Get(id)
	if err != nil {
		return Note{}, fmt.Errorf("notesService.Get:%w", err)
	}
	return note, nil
}

func (s *RepositoryImpl) List(accountID int) ([]Note, error) {
	notes, err := s.repo.List(accountID)
	if err != nil {
		return []Note{}, fmt.Errorf("notesService.List:%w", err)
	}
	return notes, nil
}

func (s *RepositoryImpl) Update(newNote UpdateNote) (int, error) {
	id, err := s.repo.Update(newNote)
	if err != nil {
		return -1, fmt.Errorf("notesService.Update:%w", err)
	}
	return id, nil
}

type PostgresNotesImpl struct{ postgresDB *pgxpool.Pool }

func PostgresNewRepository(postgresDB *pgxpool.Pool) *PostgresNotesImpl {
	return &PostgresNotesImpl{postgresDB: postgresDB}
}

func (s *PostgresNotesImpl) Create(ctx context.Context, accountID int, title string, body string) (int, error) {
	sql := `INSERT INTO note (account_id, title, body)
			VALUES ($1, $2, $3)
			RETURNING id`
	row := s.postgresDB.QueryRow(ctx, sql, accountID, title, body)
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

func (s *PostgresNotesImpl) List(accountID int) ([]Note, error) {
	panic("list not implemented yet")
}

func (s *PostgresNotesImpl) Update(newNote UpdateNote) (int, error) {
	panic("update not implemented yet")
}

func JsonValidator(note Note) error {
	if note.Title == "" {
		return fmt.Errorf("no title in json")
	}

	if note.AccountID == 0 {
		return fmt.Errorf("no accountID in json")
	}

	return nil
}
