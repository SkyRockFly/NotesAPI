package noterepository

import (
	notetype "NotesService/internal/noteType"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresImpl struct{ postgresDB *pgxpool.Pool }

func NewPostgres(postgresDB *pgxpool.Pool) *PostgresImpl {
	return &PostgresImpl{postgresDB: postgresDB}
}

func (s *PostgresImpl) Create(ctx context.Context, accountID int, title string, body string) (int, error) {
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

func (s *PostgresImpl) Delete(ctx context.Context, id int, accountID int) error {
	sql := `DELETE FROM note
			WHERE id = $1 AND account_id = $2`
	resp, err := s.postgresDB.Exec(ctx, sql, id, accountID)
	if err != nil {
		return fmt.Errorf("noteRepository.PostgresImpl.Delete:%w", err)
	}
	affected := resp.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("noteRepository.PostgresImpl.Delete:note not found" +
			"or not belong to this accountID")
	}
	return nil
}

func (s *PostgresImpl) Get(id int) (*notetype.Repo, error) {
	panic("get not implemented yet")
}

func (s *PostgresImpl) List(accountID int) ([]notetype.Repo, error) {
	panic("list not implemented yet")
}

func (s *PostgresImpl) Update(newNote notetype.Update) (int, error) {
	panic("update not implemented yet")
}
