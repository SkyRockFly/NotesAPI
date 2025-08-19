package noterepository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found") // 404

type PostgresImpl struct {
	postgresDB *pgxpool.Pool
}

func NewPostgres(postgresDB *pgxpool.Pool) *PostgresImpl {
	return &PostgresImpl{postgresDB: postgresDB}
}

func (s *PostgresImpl) Create(ctx context.Context, accountID int, title string, body string) (int, error) {
	sql := `INSERT INTO note (account_id, title, body)
			VALUES ($1, $2, $3)
			RETURNING id;`
	row := s.postgresDB.QueryRow(ctx, sql, accountID, title, body)
	var id int
	if err := row.Scan(&id); err != nil {
		return 0, fmt.Errorf("query: %w", err)
	}

	return id, nil
}

func (s *PostgresImpl) Delete(ctx context.Context, id int, accountID int) error {
	sql := `UPDATE note
			SET deleted_at = NOW()
			WHERE id = $1 AND account_id = $2`
	resp, err := s.postgresDB.Exec(ctx, sql, id, accountID)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}
	affected := resp.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("note: %w", ErrNotFound)
	}
	return nil
}

func (s *PostgresImpl) Get(ctx context.Context, id int, accountID int) (Note, error) {
	sql := `SELECT id, account_id, title, body, created_at, updated_at 
			FROM note
			WHERE id = $1 AND account_id = $2`
	row := s.postgresDB.QueryRow(ctx, sql, id, accountID)
	var note Note
	if err := row.Scan(&note.ID, &note.AccountID, &note.Title, &note.Body,
		&note.CreatedAt, &note.UpdatedAt); err != nil {
		return Note{}, fmt.Errorf("noteRepo.Get:%w : %w", err, ErrNotFound)
	}

	return note, nil
}

func (s *PostgresImpl) List(ctx context.Context, accountID int) ([]Note, error) {
	sql := `SELECT id, account_id, title, body, created_at, updated_at FROM note
			WHERE account_id = $1;`
	rows, err := s.postgresDB.Query(ctx, sql, accountID)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var note Note
		if err := rows.Scan(
			&note.ID,
			&note.AccountID,
			&note.Title,
			&note.Body,
			&note.CreatedAt,
			&note.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("next: %w", err)
		}
		notes = append(notes, note)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %w", err)
	}

	if len(notes) == 0 {
		return nil, ErrNotFound
	}

	return notes, nil
}

func (s *PostgresImpl) Update(ctx context.Context, title string, body string, id int, accountID int) error {
	sql := `UPDATE note
			SET title = $1, body = $2
			WHERE ID = $3 AND account_id = $4`
	resp, err := s.postgresDB.Exec(ctx, sql, title, body, id, accountID)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}
	affected := resp.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("note: %w", ErrNotFound)
	}
	return nil
}
