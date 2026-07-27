package refreshtokenpg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	refreshtokenrepo "notes/internal/gateway/repository/refreshToken"
	"notes/internal/pkg/apperror"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	sqlCreate = `INSERT INTO refresh_token (selector, private_hash, user_id, expired_at) 
VALUES ($1, $2, $3,
(now() AT TIME ZONE 'utc' + $4));`
	sqlRevokeAll = `UPDATE refresh_token
SET revoked = true WHERE user_id = $1;`
	sqlRevoke = `UPDATE refresh_token
SET revoked = true WHERE selector = $1;`
	sqlGet = `SELECT rt.selector,rt.private_hash,rt.user_id,rt.issued_at,rt.expired_at,rt.revoked FROM refresh_token rt
JOIN app_user au ON rt.user_id = au.id
WHERE rt.selector = $1 AND au.deleted_at IS NULL;`
)

type Repository struct {
	postgresDB *pgxpool.Pool
}

func NewRepository(postgresDB *pgxpool.Pool) *Repository {
	return &Repository{postgresDB: postgresDB}
}

func (s *Repository) Create(ctx context.Context, token refreshtokenrepo.CreateReq) error {
	_, err := s.postgresDB.Exec(
		ctx,
		sqlCreate,
		token.Selector,
		token.Private,
		token.UserID,
		token.TTL,
	)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	return nil
}

func (s *Repository) RevokeAll(ctx context.Context, userID int) error {
	resp, err := s.postgresDB.Exec(ctx, sqlRevokeAll, userID)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}
	if resp.RowsAffected() == 0 {
		return fmt.Errorf("rows: %w", apperror.ErrNotFound)
	}
	return nil
}

func (s *Repository) Revoke(ctx context.Context, selector string) error {
	resp, err := s.postgresDB.Exec(ctx, sqlRevoke, selector)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}
	if resp.RowsAffected() == 0 {
		return fmt.Errorf("rows: %w", apperror.ErrNotFound)
	}
	return nil
}

func (s *Repository) Get(ctx context.Context, selector string) (refreshtokenrepo.Token, error) {
	row := s.postgresDB.QueryRow(ctx, sqlGet, selector)
	var token refreshtokenrepo.Token
	if err := row.Scan(
		&token.Selector,
		&token.PrivateHash,
		&token.UserID,
		&token.IssuedAt,
		&token.ExpiredAt,
		&token.Revoked,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return refreshtokenrepo.Token{}, fmt.Errorf("scan: %w", apperror.ErrNotFound)
		}
		return refreshtokenrepo.Token{}, fmt.Errorf("scan: %w ", err)
	}

	return token, nil
}
