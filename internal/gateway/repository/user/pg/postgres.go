package userpg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	userrepo "notes/internal/gateway/repository/user"
	"notes/internal/pkg/apperror"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	sqlCreate = `INSERT INTO app_user (login, password, email) 
VALUES ($1, $2, $3)
RETURNING id;`
	sqlDelete     = `UPDATE FROM app_user WHERE id = $1;`
	sqlGetByLogin = `SELECT id, login, password, email FROM app_user
WHERE login = $1;`

	pgUniqueErrorCode = "23505"
)

type Repository struct {
	postgresDB *pgxpool.Pool
}

func NewRepository(postgresDB *pgxpool.Pool) *Repository {
	return &Repository{postgresDB: postgresDB}
}

func (s *Repository) Create(ctx context.Context, user userrepo.CreateUserReq) (int, error) {
	var id int
	err := s.postgresDB.QueryRow(
		ctx,
		sqlCreate,
		strings.ToLower(strings.TrimSpace(user.Login)),
		user.Password,
		strings.ToLower(strings.TrimSpace(user.Email)),
	).Scan(&id)
	if err != nil {
		var pg *pgconn.PgError
		if errors.As(err, &pg) && pg.Code == pgUniqueErrorCode {
			return 0, fmt.Errorf("query: %w", apperror.ErrAlreadyExists)
		}
		return 0, fmt.Errorf("query: %w", err)
	}
	return id, nil
}

func (s *Repository) Delete(ctx context.Context, id int) error {
	resp, err := s.postgresDB.Exec(ctx, sqlDelete, id)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}
	if resp.RowsAffected() == 0 {
		return fmt.Errorf("rows affected: %w", apperror.ErrNotFound)
	}
	return nil
}

func (s *Repository) Get(ctx context.Context, login string) (userrepo.User, error) {
	row := s.postgresDB.QueryRow(
		ctx,
		sqlGetByLogin,
		strings.ToLower(strings.TrimSpace(login)))
	var user userrepo.User
	if err := row.Scan(
		&user.ID,
		&user.Login,
		&user.Password,
		&user.Email,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return userrepo.User{}, fmt.Errorf("scan: %w", apperror.ErrNotFound)
		}
		return userrepo.User{}, fmt.Errorf("scan: %w ", err)
	}

	return user, nil
}
