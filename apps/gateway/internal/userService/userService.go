package userservice

import (
	"context"
	"errors"
	"fmt"
	usernoterepo "notes/apps/gateway/internal/userNoteRepo"
	"time"
)

var ErrInvalid = errors.New("invalid") // 400

type IRepository interface {
	Get(ctx context.Context, id, accountID int) (usernoterepo.UserNote, error)
	Create(ctx context.Context, accountID int, title, body string) (usernoterepo.UserNote, error)
	Update(ctx context.Context, id, accountID int, title, body string) (usernoterepo.UserNote, error)
	List(ctx context.Context, accountID int) ([]usernoterepo.UserNote, error)
	Delete(ctx context.Context, id, accountID int) (usernoterepo.UserNote, error)
}

type Service struct {
	repo IRepository
}

type Note struct {
	ID        int
	AccountID int
	Title     string
	Body      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewService(r IRepository) *Service {
	return &Service{
		repo: r,
	}
}

func (s *Service) Get(ctx context.Context, note Note) (Note, error) {
	if note.ID < 0 {
		return Note{}, fmt.Errorf("%w : id less than zero", ErrInvalid)
	}
	if note.AccountID < 0 {
		return Note{}, fmt.Errorf("%w : accountID less than zero", ErrInvalid)
	}

	repoNote, err := s.repo.Get(ctx, note.ID, note.AccountID)
	if err != nil {
		return Note{}, fmt.Errorf("userNoteRepo.Get: %w", err)
	}

	svcNote := Note{
		ID:        repoNote.ID,
		AccountID: repoNote.AccountID,
		Title:     repoNote.Title,
		Body:      repoNote.Body,
		CreatedAt: repoNote.CreatedAt,
		UpdatedAt: repoNote.UpdatedAt,
	}

	return svcNote, nil
}
