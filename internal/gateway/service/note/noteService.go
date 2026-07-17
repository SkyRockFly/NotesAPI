package notesvc

import (
	"context"
	"fmt"
	noterepo "notes/internal/gateway/repository/note"
	"notes/internal/pkg/apperror"
	"notes/internal/pkg/kit"
	"time"

	"github.com/go-playground/validator/v10"
)

const (
	maxTitleSize = 255
)

type Service struct {
	repo     noterepo.INote
	validate *validator.Validate
}

type Note struct {
	ID        int64
	AccountID int
	Title     string
	Body      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CreateReq struct {
	AccountID int    `validate:"min=1"`
	Title     string `validate:"required,max=255"`
	Body      string `validate:"required"`
}

type UpdateReq struct {
	ID        int64  `validate:"min=1"`
	AccountID int    `validate:"min=1"`
	Title     string `validate:"required,max=255"`
	Body      string `validate:"required"`
}

type ListReq struct {
	AccountID int   `validate:"min=1"`
	Limit     int   `validate:"min=1"`
	Cursor    int64 `validate:"min=0"`
	Next      bool
}

type ListResp struct {
	CursorNext int64
	CursorPrev int64
	Notes      []Note
	HasMore    bool
}

type DeleteReq struct {
	ID        int64 `validate:"min=1"`
	AccountID int   `validate:"min=1"`
}

type GetReq struct {
	ID        int64 `validate:"min=1"`
	AccountID int   `validate:"min=1"`
}

func NewService(r noterepo.INote) *Service {
	return &Service{
		repo:     r,
		validate: validator.New(),
	}
}

func (s *Service) Create(ctx context.Context, req CreateReq) (int64, error) {
	if err := kit.ValidateStruct(s.validate, req); err != nil {
		return 0, fmt.Errorf("%w , validate struct: %w", apperror.ErrBadRequest, err)
	}

	createReq := noterepo.CreateReq{
		AccountID: req.AccountID,
		Title:     req.Title,
		Body:      req.Body,
	}

	id, err := s.repo.Create(ctx, createReq)
	if err != nil {
		return 0, fmt.Errorf("userNoteRepo.Get: %w", err)
	}

	return id, nil
}

func (s *Service) Delete(ctx context.Context, req DeleteReq) error {
	if err := kit.ValidateStruct(s.validate, req); err != nil {
		return fmt.Errorf("%w , validate struct: %w", apperror.ErrBadRequest, err)
	}

	deleteReq := noterepo.DeleteReq{
		ID:        req.ID,
		AccountID: req.AccountID,
	}

	if err := s.repo.Delete(ctx, deleteReq); err != nil {
		return fmt.Errorf("userNoteRepo.Delete: %w", err)
	}

	return nil
}

func (s *Service) Update(ctx context.Context, req UpdateReq) error {
	if err := kit.ValidateStruct(s.validate, req); err != nil {
		return fmt.Errorf("%w , validate struct: %w", apperror.ErrBadRequest, err)
	}

	updateReq := noterepo.UpdateReq{
		ID:        req.ID,
		AccountID: req.AccountID,
		Title:     req.Title,
		Body:      req.Body,
	}

	if err := s.repo.Update(ctx, updateReq); err != nil {
		return fmt.Errorf("userNoteRepo.Update: %w", err)
	}

	return nil
}

func (s *Service) List(ctx context.Context, req ListReq) (ListResp, error) {
	if err := kit.ValidateStruct(s.validate, req); err != nil {
		return ListResp{}, fmt.Errorf("%w , validate struct: %w", apperror.ErrBadRequest, err)
	}

	listReq := noterepo.ListReq{
		AccountID: req.AccountID,
		Cursor:    req.Cursor,
		Limit:     req.Limit,
		Next:      req.Next,
	}

	repoResp, err := s.repo.List(ctx, listReq)
	if err != nil {
		return ListResp{}, fmt.Errorf("userNoteRepo.List: %w", err)
	}

	svcNotes := make([]Note, 0, len(repoResp.Notes))
	var svcNote Note
	for _, repoNote := range repoResp.Notes {
		svcNote = mapToSvcNote(repoNote)
		svcNotes = append(svcNotes, svcNote)
	}

	resp := ListResp{
		CursorNext: repoResp.CursorNext,
		CursorPrev: repoResp.CursorPrev,
		Notes:      svcNotes,
		HasMore:    repoResp.HasMore,
	}

	return resp, nil
}

func (s *Service) Get(ctx context.Context, req GetReq) (Note, error) {
	if err := kit.ValidateStruct(s.validate, req); err != nil {
		return Note{}, fmt.Errorf("%w , validate struct: %w", apperror.ErrBadRequest, err)
	}

	getReq := noterepo.GetReq{
		ID:        req.ID,
		AccountID: req.AccountID,
	}

	repoNote, err := s.repo.Get(ctx, getReq)
	if err != nil {
		return Note{}, fmt.Errorf("userNoteRepo.Get: %w", err)
	}

	svcNote := mapToSvcNote(repoNote)

	return svcNote, nil
}

func mapToSvcNote(repoNote noterepo.Note) Note {
	return Note{
		ID:        repoNote.ID,
		AccountID: repoNote.AccountID,
		Title:     repoNote.Title,
		Body:      repoNote.Body,
		CreatedAt: repoNote.CreatedAt,
		UpdatedAt: repoNote.UpdatedAt,
	}
}
