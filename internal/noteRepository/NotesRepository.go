package noterepository

import (
	notetype "NotesService/internal/noteType"
	"context"
)

type Repository interface {
	Create(ctx context.Context, accountID int, title string, body string) (int, error)
	Delete(ctx context.Context, id int, accountID int) error
	Get(id int) (*notetype.Repo, error)
	List(accountID int) ([]notetype.Repo, error)
	Update(newNote notetype.Update) (int, error)
}
