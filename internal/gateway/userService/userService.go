package userservice

import (
	"context"
	"errors"
	"fmt"
	usernoterepo "notes/internal/gateway/userNoteRepo"
	"strings"
	"time"
)

var (
	ErrInvalid = errors.New("invalid")      // 400
	ErrServer  = errors.New("server error") //500
)

type IRepository interface {
	Get(ctx context.Context, id, accountID int) (usernoterepo.UserNote, error)
	Create(ctx context.Context, accountID int, title, body string) (int, error)
	Update(ctx context.Context, id, accountID int, title, body string) (bool, error)
	List(ctx context.Context, accountID int) ([]usernoterepo.UserNote, error)
	Delete(ctx context.Context, id, accountID int) (bool, error)
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

type CreateNoteData struct {
	AccountID int
	Title     string
	Body      string
}

type UpdateNoteData struct {
	ID        int
	AccountID int
	Title     string
	Body      string
}

type ListNoteData struct {
	AccountID int
}

type DeleteNoteData struct {
	ID        int
	AccountID int
}

type GetNoteData struct {
	ID        int
	AccountID int
}

func NewService(r IRepository) *Service {
	return &Service{
		repo: r,
	}
}

func (s *Service) Create(ctx context.Context, note CreateNoteData) (int, error) {
	if note.AccountID < 1 {
		return 0, fmt.Errorf("%w :accountID less than 1", ErrInvalid)
	}
	if strings.TrimSpace(note.Title) == "" {
		return 0, fmt.Errorf("%w : empty title", ErrInvalid)
	}

	if len(strings.TrimSpace(note.Title)) > 255 {
		return 0, fmt.Errorf("%w : title more than 255 chars", ErrInvalid)
	}

	id, err := s.repo.Create(ctx, note.AccountID, note.Title, note.Body)
	if err != nil {
		return 0, fmt.Errorf("userNoteRepo.Get: %w", err)
	}
	if id < 1 {
		return 0, fmt.Errorf("userNoteRepo.Get: %w - id from server less than 1", ErrServer)
	}

	return id, nil
}

func (s *Service) Delete(ctx context.Context, note DeleteNoteData) (bool, error) {
	if note.ID < 1 {
		return false, fmt.Errorf("%w :id less than 1", ErrInvalid)
	}

	if note.AccountID < 1 {
		return false, fmt.Errorf("%w :accountID less than 1", ErrInvalid)
	}

	deleted, err := s.repo.Delete(ctx, note.ID, note.AccountID)
	if err != nil {
		return false, fmt.Errorf("userNoteRepo.Delete: %w", err)
	}

	return deleted, nil
}

func (s *Service) Update(ctx context.Context, note UpdateNoteData) (bool, error) {
	if note.ID < 1 {
		return false, fmt.Errorf("%w :id less than 1", ErrInvalid)
	}

	if note.AccountID < 1 {
		return false, fmt.Errorf("%w :accountID less than 1", ErrInvalid)
	}

	if strings.TrimSpace(note.Title) == "" {
		return false, fmt.Errorf("%w : empty title", ErrInvalid)
	}

	if len(strings.TrimSpace(note.Title)) > 255 {
		return false, fmt.Errorf("%w : title more than 255 chars", ErrInvalid)
	}

	updated, err := s.repo.Update(ctx, note.ID, note.AccountID, note.Title, note.Body)
	if err != nil {
		return false, fmt.Errorf("userNoteRepo.Delete: %w", err)
	}

	return updated, nil
}

func (s *Service) List(ctx context.Context, note ListNoteData) ([]Note, error) {
	if note.AccountID < 1 {
		return nil, fmt.Errorf("%w :accountID less than 1", ErrInvalid)
	}

	repoNotes, err := s.repo.List(ctx, note.AccountID)
	if err != nil {
		return nil, fmt.Errorf("userNoteRepo.Delete: %w", err)
	}

	svcNotes := make([]Note, 0, len(repoNotes))
	var svcNote Note
	for _, repoNote := range repoNotes {
		svcNote = mapToSvcNote(repoNote)
		svcNotes = append(svcNotes, svcNote)
	}

	return svcNotes, nil
}

func (s *Service) Get(ctx context.Context, note GetNoteData) (Note, error) {
	if note.ID < 1 {
		return Note{}, fmt.Errorf("%w : id less than 1", ErrInvalid)
	}
	if note.AccountID < 1 {
		return Note{}, fmt.Errorf("%w : accountID less than 1", ErrInvalid)
	}

	repoNote, err := s.repo.Get(ctx, note.ID, note.AccountID)
	if err != nil {
		return Note{}, fmt.Errorf("userNoteRepo.Get: %w", err)
	}

	svcNote := mapToSvcNote(repoNote)

	return svcNote, nil
}

func mapToSvcNote(repoNote usernoterepo.UserNote) Note {
	return Note{
		ID:        repoNote.ID,
		AccountID: repoNote.AccountID,
		Title:     repoNote.Title,
		Body:      repoNote.Body,
		CreatedAt: repoNote.CreatedAt,
		UpdatedAt: repoNote.UpdatedAt,
	}
}
