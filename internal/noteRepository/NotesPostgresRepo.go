package noterepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found") // 404
const (
	sqlCreate = `INSERT INTO note (account_id, title, body)
VALUES ($1, $2, $3)
RETURNING id;`
	sqlDelete = `UPDATE note
SET deleted_at = NOW()
WHERE id = $1 AND account_id = $2;`
	sqlGet = `SELECT id, account_id, title, body, created_at, updated_at 
FROM note
WHERE id = $1 AND account_id = $2;`
	sqlList = `SELECT id, account_id, title, body, created_at, updated_at FROM note
WHERE account_id = $1;`
	sqlUpdate = `UPDATE note
SET title = $1, body = $2
WHERE ID = $3 AND account_id = $4;`
)

type PostgresImpl struct {
	postgresDB *pgxpool.Pool
}

func NewPostgres(postgresDB *pgxpool.Pool) *PostgresImpl {
	return &PostgresImpl{postgresDB: postgresDB}
}

func (s *PostgresImpl) Create(ctx context.Context, accountID int, title string, body string) (int, error) {
	row := s.postgresDB.QueryRow(ctx, sqlCreate, accountID, title, body)
	var id int
	if err := row.Scan(&id); err != nil {
		return 0, fmt.Errorf("query: %w", err)
	}

	return id, nil
}

func (s *PostgresImpl) Delete(ctx context.Context, id int, accountID int) error {
	resp, err := s.postgresDB.Exec(ctx, sqlDelete, id, accountID)
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
	row := s.postgresDB.QueryRow(ctx, sqlGet, id, accountID)
	var note Note
	if err := row.Scan(
		&note.ID,
		&note.AccountID,
		&note.Title,
		&note.Body,
		&note.CreatedAt,
		&note.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Note{}, fmt.Errorf("scan: %w", ErrNotFound)
		}
		return Note{}, fmt.Errorf("scan: %w ", err)
	}

	return note, nil
}

func (s *PostgresImpl) List(ctx context.Context, accountID int) ([]Note, error) {
	rows, err := s.postgresDB.Query(ctx, sqlList, accountID)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	var notes []Note
	var note Note
	for rows.Next() {
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
		return nil, fmt.Errorf("note: %w", ErrNotFound)
	}

	return notes, nil
}

func (s *PostgresImpl) Update(ctx context.Context, title string, body string, id int, accountID int) error {
	resp, err := s.postgresDB.Exec(ctx, sqlUpdate, title, body, id, accountID)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}
	affected := resp.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("note: %w", ErrNotFound)
	}
	return nil
}
