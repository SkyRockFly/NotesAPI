package noteservice

import (
	noterepository "NotesService/internal/noteRepository"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrInvalid = errors.New("invalid") // 400

type IRepository interface {
	Create(ctx context.Context, accountID int, title string, body string) (int, error)
	Delete(ctx context.Context, id, accountID int) error
	Get(ctx context.Context, id, accountID int) (noterepository.Note, error)
	List(ctx context.Context, accountID int) ([]noterepository.Note, error)
	Update(ctx context.Context, title, body string, id, accountID int) error
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
	DeletedAt time.Time
}

// t, _ := time.Parse(time.DateTime, "")
//	t.IsZero()

func NewService(repo IRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, note Note) (int, error) {
	if note.AccountID < 1 {
		return 0, fmt.Errorf("%w: ID cannot be less than 1", ErrInvalid)
	}
	if strings.TrimSpace(note.Title) == "" {
		return 0, fmt.Errorf("%w: no title", ErrInvalid)
	}
	if len(note.Title) > 255 {
		return 0, fmt.Errorf("%w: the length of title is more than 255 symbols", ErrInvalid)
	}

	id, err := s.repo.Create(ctx, note.AccountID, note.Title, note.Body)
	if err != nil {
		return 0, fmt.Errorf("noteRepo.Create: %w", err)
	}
	return id, nil
}

func (s *Service) Delete(ctx context.Context, note Note) error {
	if note.ID < 1 {
		return fmt.Errorf("%w: ID cannot be less than 1", ErrInvalid)
	}
	if note.AccountID < 1 {
		return fmt.Errorf("%w: ID cannot be less than 1", ErrInvalid)
	}

	if err := s.repo.Delete(ctx, note.ID, note.AccountID); err != nil {
		return fmt.Errorf("noteService.Delete:%w", err)
	}
	return nil
}

func (s *Service) Get(ctx context.Context, note Note) (Note, error) {
	if note.ID < 1 {
		return Note{}, fmt.Errorf("%w: ID cannot be less than 1", ErrInvalid)
	}
	if note.AccountID < 1 {
		return Note{}, fmt.Errorf("%w: ID cannot be less than 1", ErrInvalid)
	}

	noteRepo, err := s.repo.Get(ctx, note.ID, note.AccountID)
	if err != nil {
		return Note{}, fmt.Errorf("noteService.Get:%w", err)
	}

	noteServ := repoToSVC(noteRepo)
	return noteServ, nil
}

func (s *Service) List(ctx context.Context, note Note) ([]Note, error) {
	if note.AccountID < 1 {
		return []Note{}, fmt.Errorf("%w: ID cannot be less than 1", ErrInvalid)
	}
	notes, err := s.repo.List(ctx, note.AccountID)
	if err != nil {
		return []Note{}, fmt.Errorf("notesService.List:%w", err)
	}

	servNotes := make([]Note, 0, len(notes))
	for _, noteRepo := range notes {
		note := repoToSVC(noteRepo)
		servNotes = append(servNotes, note)
	}

	return servNotes, nil
}

func (s *Service) Update(ctx context.Context, note Note) error {
	if note.ID < 1 {
		return fmt.Errorf("%w: ID cannot be less than 1", ErrInvalid)
	}
	if note.AccountID < 1 {
		return fmt.Errorf("%w: ID cannot be less than 1", ErrInvalid)
	}
	if strings.TrimSpace(note.Title) == "" {
		return fmt.Errorf("%w: no title", ErrInvalid)
	}
	if len(note.Title) > 255 {
		return fmt.Errorf("%w: the length of title is more than 255 symbols", ErrInvalid)
	}

	if err := s.repo.Update(ctx, note.Title, note.Body, note.ID, note.AccountID); err != nil {
		return fmt.Errorf("notesService.Update:%w", err)
	}
	return nil
}

func repoToSVC(noteRepo noterepository.Note) Note {
	note := Note{
		ID:        noteRepo.ID,
		AccountID: noteRepo.AccountID,
		Title:     noteRepo.Title,
		Body:      noteRepo.Body,
		CreatedAt: noteRepo.CreatedAt,
		UpdatedAt: noteRepo.UpdatedAt,
	}
	return note
}
