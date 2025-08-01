package noteservice

import (
	noterepository "NotesService/internal/noteRepository"
	notetype "NotesService/internal/noteType"
	"context"
	"fmt"
	"strings"
)

type Service struct {
	repo *noterepository.PostgresImpl
}

func NewService(repo *noterepository.PostgresImpl) *Service {
	return &Service{repo: repo}
}

func RemapDTOtoServ(noteDTO notetype.DTO) *notetype.Service {
	serviceNote := &notetype.Service{
		ID:        noteDTO.ID,
		AccountID: noteDTO.AccountID,
		Title:     noteDTO.Title,
		Body:      noteDTO.Body,
	}
	return serviceNote
}

func RemapServToRepo(serviceNote notetype.Service) *notetype.Repo {
	repositoryNote := &notetype.Repo{
		ID:        serviceNote.ID,
		AccountID: serviceNote.AccountID,
		Title:     serviceNote.Title,
		Body:      serviceNote.Body,
	}
	return repositoryNote
}

func (s *Service) Create(ctx context.Context, accountID int, title string, body string) (int, error) {
	id, err := s.repo.Create(ctx, accountID, title, body)
	if err != nil {
		return -1, fmt.Errorf("notesService.Create: %w", err)
	}
	return id, nil
}

func (s *Service) Delete(ctx context.Context, id int, accountID int) error {
	if err := s.repo.Delete(ctx, id, accountID); err != nil {
		return fmt.Errorf("notesService.Delete:%w", err)
	}
	return nil
}

func (s *Service) Get(id int) (*notetype.Repo, error) {
	note, err := s.repo.Get(id)
	if err != nil {
		return &notetype.Repo{}, fmt.Errorf("notesService.Get:%w", err)
	}
	return note, nil
}

func (s *Service) List(accountID int) ([]notetype.Repo, error) {
	notes, err := s.repo.List(accountID)
	if err != nil {
		return []notetype.Repo{}, fmt.Errorf("notesService.List:%w", err)
	}
	return notes, nil
}

func (s *Service) Update(newNote notetype.Update) (int, error) {
	id, err := s.repo.Update(newNote)
	if err != nil {
		return -1, fmt.Errorf("notesService.Update:%w", err)
	}
	return id, nil
}

func noteValidator(note notetype.Service) error {
	if strings.TrimSpace(note.Title) == "" {
		return fmt.Errorf("no title in json")
	}

	if note.AccountID <= 0 {
		return fmt.Errorf("no accountID in json")
	}

	return nil
}
