package noteservice

import (
	noterepository "NotesService/internal/noteRepository"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalid = errors.New("invalid") // 400
)

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

//t, _ := time.Parse(time.DateTime, "")
//	t.IsZero()

func (n *Note) CreateValidator() error { //без методов Note
	if n.AccountID < 1 {
		return fmt.Errorf("%w: ID cannot be less than 1", ErrInvalid)
	}
	if strings.TrimSpace(n.Title) == "" {
		return fmt.Errorf("%w: no title", ErrInvalid)
	}
	if len(n.Title) > 255 {
		return fmt.Errorf("%w: the length of title is more than 255 symbols", ErrInvalid)
	}
	return nil
}

func (n *Note) DeleteValidator() error { //Validate
	if n.ID < 1 {
		return fmt.Errorf("%w: ID cannot be less than 1", ErrInvalid)
	}
	if n.AccountID < 1 {
		return fmt.Errorf("%w: ID cannot be less than 1", ErrInvalid)
	}
	return nil
}

func (n *Note) GetValidator() error {
	if n.ID < 1 {
		return fmt.Errorf("%w: ID cannot be less than 1", ErrInvalid)
	}
	if n.AccountID < 1 {
		return fmt.Errorf("%w: ID cannot be less than 1", ErrInvalid)
	}
	return nil
}

func (n *Note) UpdateValidator() error {
	if n.ID < 1 {
		return fmt.Errorf("%w: ID cannot be less than 1", ErrInvalid)
	}
	if n.AccountID < 1 {
		return fmt.Errorf("%w: ID cannot be less than 1", ErrInvalid)
	}
	if strings.TrimSpace(n.Title) == "" {
		return fmt.Errorf("%w: no title", ErrInvalid)
	}
	if len(n.Title) > 255 {
		return fmt.Errorf("%w: the length of title is more than 255 symbols", ErrInvalid)
	}
	return nil
}

func (n *Note) ListValidator() error {
	if n.AccountID < 1 {
		return fmt.Errorf("%w: ID cannot be less than 1", ErrInvalid)
	}
	return nil
}

func NewService(repo IRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, note *Note) (int, error) {
	if err := note.CreateValidator(); err != nil {
		return 0, fmt.Errorf("validator: %w", err)
	}

	id, err := s.repo.Create(ctx, note.AccountID, note.Title, note.Body)
	if err != nil {
		return 0, fmt.Errorf("noteRepo.Create: %w", err)
	}
	return id, nil
}

func (s *Service) Delete(ctx context.Context, note *Note) error {
	if err := note.DeleteValidator(); err != nil {
		return fmt.Errorf("validator: %w", err)
	}

	if err := s.repo.Delete(ctx, note.ID, note.AccountID); err != nil {
		return fmt.Errorf("noteService.Delete:%w", err)
	}
	return nil
}

func (s *Service) Get(ctx context.Context, note *Note) (Note, error) {
	if err := note.GetValidator(); err != nil {
		return Note{}, fmt.Errorf("validator: %w", err)
	}

	noteRepo, err := s.repo.Get(ctx, note.ID, note.AccountID)
	if err != nil {
		return Note{}, fmt.Errorf("noteService.Get:%w", err)
	}

	noteServ := repoToServ(noteRepo)
	return noteServ, nil
}

func (s *Service) List(ctx context.Context, note *Note) ([]Note, error) {
	if err := note.ListValidator(); err != nil {
		return nil, fmt.Errorf("validator: %w", err)
	}

	notes, err := s.repo.List(ctx, note.AccountID)
	if err != nil {
		return []Note{}, fmt.Errorf("notesService.List:%w", err)
	}

	servNotes := make([]Note, 0, len(notes))
	for _, noteRepo := range notes {
		note := repoToServ(noteRepo)
		servNotes = append(servNotes, note)
	}

	return servNotes, nil
}

func (s *Service) Update(ctx context.Context, note *Note) error {
	if err := note.UpdateValidator(); err != nil {
		return fmt.Errorf("validator:%w", err)
	}

	if err := s.repo.Update(ctx, note.Title, note.Body, note.ID, note.AccountID); err != nil {
		return fmt.Errorf("notesService.Update:%w", err)
	}
	return nil
}

func repoToServ(noteRepo noterepository.Note) Note {
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
