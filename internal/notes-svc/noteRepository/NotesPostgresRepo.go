package noterepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"notes/internal/pkg/apperror"
	"slices"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	sqlCreate = `INSERT INTO note (account_id, title, body)
VALUES ($1, $2, $3)
RETURNING id;`
	sqlDelete = `UPDATE note
SET deleted_at = (now() AT TIME ZONE 'UTC'),
updated_at = (now() AT TIME ZONE 'UTC')
WHERE id = $1 AND account_id = $2 AND deleted_at IS NULL;`
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

func (s *PostgresImpl) Create(ctx context.Context, req CreateReq) (int, error) {
	row := s.postgresDB.QueryRow(ctx, sqlCreate, req.AccountID, req.Title, req.Body)
	var id int
	if err := row.Scan(&id); err != nil {
		return 0, fmt.Errorf("query: %w", err)
	}

	return id, nil
}

func (s *PostgresImpl) Delete(ctx context.Context, req DeleteReq) error {
	resp, err := s.postgresDB.Exec(ctx, sqlDelete, req.ID, req.AccountID)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}
	affected := resp.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("note: %w", apperror.ErrNotFound)
	}
	return nil
}

func (s *PostgresImpl) Get(ctx context.Context, req GetReq) (Note, error) {
	row := s.postgresDB.QueryRow(ctx, sqlGet, req.ID, req.AccountID)
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
			return Note{}, fmt.Errorf("scan: %w", apperror.ErrNotFound)
		}
		return Note{}, fmt.Errorf("scan: %w ", err)
	}

	return note, nil
}

func (s *PostgresImpl) List(ctx context.Context, req ListReq) (ListResp, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	q := psql.
		Select("id, account_id, title, body, created_at, updated_at").
		From("note").
		Where("account_id = ?", req.AccountID)

	limitPlusOne := req.Limit + 1
	if req.Next {
		q = q.Where("id > ?", req.Cursor)
		q = q.OrderBy("id ASC").Limit(uint64(limitPlusOne))
	} else {
		q = q.Where("id < ?", req.Cursor)
		q = q.OrderBy("id DESC").Limit(uint64(limitPlusOne))
	}

	sqlStr, args, err := q.ToSql()
	if err != nil {
		return ListResp{}, fmt.Errorf("build sql: %w", err)
	}
	rows, err := s.postgresDB.Query(ctx, sqlStr, args...)
	if err != nil {
		return ListResp{}, fmt.Errorf("query: %w", err)
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
			return ListResp{}, fmt.Errorf("next: %w", err)
		}
		notes = append(notes, note)
	}
	if err := rows.Err(); err != nil {
		return ListResp{}, fmt.Errorf("rows: %w", err)
	}

	n := len(notes)
	if n == 0 {
		return ListResp{}, fmt.Errorf("notes: %w", apperror.ErrNotFound)
	}

	hasMore := len(notes) > req.Limit
	if hasMore {
		notes = notes[:req.Limit]
	}
	if !req.Next {
		slices.Reverse(notes)
	}
	top := notes[len(notes)-1]
	bottom := notes[0]

	resp := ListResp{
		Notes:      notes,
		CursorNext: top.ID,
		CursorPrev: bottom.ID,
		HasMore:    hasMore,
	}

	return resp, nil
}

func (s *PostgresImpl) Update(ctx context.Context, req UpdateReq) error {
	resp, err := s.postgresDB.Exec(ctx, sqlUpdate, req.Title, req.Body, req.ID, req.AccountID)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}
	affected := resp.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("note: %w", apperror.ErrNotFound)
	}
	return nil
}
