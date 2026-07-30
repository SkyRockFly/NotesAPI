package noteservice

import (
	"context"
	"fmt"
	noterepository "notes/internal/notes-svc/noteRepository"
	"notes/internal/pkg/apperror"
	"notes/internal/pkg/kit"
	"time"

	"github.com/go-playground/validator/v10"
)

type Service struct {
	repo     noterepository.IRepository
	validate *validator.Validate
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

type ListReq struct {
	AccountID int `validate:"min=1"`
	Limit     int `validate:"min=1"`
	Cursor    int `validate:"min=0"`
	Next      bool
}

type ListResp struct {
	CursorNext int
	CursorPrev int
	Notes      []Note
	HasMore    bool
}

type GetReq struct {
	ID        int `validate:"min=1"`
	AccountID int `validate:"min=1"`
}

type CreateReq struct {
	AccountID int    `validate:"min=1"`
	Title     string `validate:"required,max=255"`
	Body      string `validate:"required"`
}

type DeleteReq struct {
	ID        int `validate:"min=1"`
	AccountID int `validate:"min=1"`
}

type UpdateReq struct {
	ID        int    `validate:"min=1"`
	AccountID int    `validate:"min=1"`
	Title     string `validate:"required,max=255"`
	Body      string
}

func NewService(repo noterepository.IRepository) *Service {
	return &Service{
		repo:     repo,
		validate: initValidator(),
	}
}

func (s *Service) Create(ctx context.Context, req CreateReq) (int, error) {
	if err := kit.ValidateStruct(s.validate, req); err != nil {
		return 0, fmt.Errorf("%w : validate struct: %w", apperror.ErrBadRequest, err)
	}

	createReq := noterepository.CreateReq{
		AccountID: req.AccountID,
		Title:     req.Title,
		Body:      req.Body,
	}

	id, err := s.repo.Create(ctx, createReq)
	if err != nil {
		return 0, fmt.Errorf("noteRepo.Create: %w", err)
	}
	return id, nil
}

func (s *Service) Delete(ctx context.Context, req DeleteReq) error {
	if err := kit.ValidateStruct(s.validate, req); err != nil {
		return fmt.Errorf("%w : validate struct: %w", apperror.ErrBadRequest, err)
	}

	deleteReq := noterepository.DeleteReq{
		AccountID: req.AccountID,
		ID:        req.ID,
	}

	if err := s.repo.Delete(ctx, deleteReq); err != nil {
		return fmt.Errorf("noteService.Delete:%w", err)
	}
	return nil
}

func (s *Service) Get(ctx context.Context, req GetReq) (Note, error) {
	if err := kit.ValidateStruct(s.validate, req); err != nil {
		return Note{}, fmt.Errorf("%w : validate struct: %w", apperror.ErrBadRequest, err)
	}

	getReq := noterepository.GetReq{
		AccountID: req.AccountID,
		ID:        req.ID,
	}

	noteRepo, err := s.repo.Get(ctx, getReq)
	if err != nil {
		return Note{}, fmt.Errorf("noteService.Get:%w", err)
	}

	noteServ := repoToSVC(noteRepo)
	return noteServ, nil
}

func (s *Service) List(ctx context.Context, req ListReq) (ListResp, error) {
	if err := kit.ValidateStruct(s.validate, req); err != nil {
		return ListResp{}, fmt.Errorf("%w : validate struct: %w", apperror.ErrBadRequest, err)
	}

	listReq := noterepository.ListReq{
		AccountID: req.AccountID,
		Limit:     req.Limit,
		Cursor:    req.Cursor,
		Next:      req.Next,
	}
	repoResp, err := s.repo.List(ctx, listReq)
	if err != nil {
		return ListResp{}, fmt.Errorf("notesService.List:%w", err)
	}

	servNotes := make([]Note, 0, len(repoResp.Notes))
	for _, noteRepo := range repoResp.Notes {
		note := repoToSVC(noteRepo)
		servNotes = append(servNotes, note)
	}

	resp := ListResp{
		CursorNext: repoResp.CursorNext,
		CursorPrev: repoResp.CursorPrev,
		HasMore:    repoResp.HasMore,
		Notes:      servNotes,
	}

	return resp, nil
}

func (s *Service) Update(ctx context.Context, req UpdateReq) error {
	if err := kit.ValidateStruct(s.validate, req); err != nil {
		return fmt.Errorf("%w : validate struct: %w", apperror.ErrBadRequest, err)
	}

	updateReq := noterepository.UpdateReq{
		ID:        req.ID,
		AccountID: req.AccountID,
		Title:     req.Title,
		Body:      req.Body,
	}

	if err := s.repo.Update(ctx, updateReq); err != nil {
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

func initValidator() *validator.Validate {
	v := validator.New()

	v.RegisterStructValidation(validateListImagesReq, ListReq{})

	return v
}

func validateListImagesReq(sl validator.StructLevel) {
	req := sl.Current().Interface().(ListReq)

	hasId := req.Cursor > 0

	if !req.Next && !hasId {
		sl.ReportError(req.Next, "Next", "next", "cursorpair", "")
	}
}
