package notemock

import (
	"context"
	notesv1 "notes/internal/grpc/generated"
)

type FakeNotesServer struct {
	notesv1.UnimplementedNotesServiceServer

	CreateNoteFn func(ctx context.Context,
		req *notesv1.CreateNoteReq) (*notesv1.CreateNoteResp, error)

	GetNoteFn func(ctx context.Context,
		req *notesv1.GetNoteReq) (*notesv1.GetNoteResp, error)

	UpdateNoteFn func(ctx context.Context,
		req *notesv1.UpdateNoteReq) (*notesv1.UpdateNoteResp, error)

	DeleteNoteFn func(ctx context.Context, req *notesv1.DeleteNoteReq,
	) (*notesv1.DeleteNoteResp, error)

	ListNoteFn func(ctx context.Context, req *notesv1.ListNoteReq,
	) (*notesv1.ListNoteResp, error)
}

func (s *FakeNotesServer) CreateNote(
	ctx context.Context,
	req *notesv1.CreateNoteReq,
) (*notesv1.CreateNoteResp, error) {
	return s.CreateNoteFn(ctx, req)
}

func (s *FakeNotesServer) GetNote(
	ctx context.Context,
	req *notesv1.GetNoteReq,
) (*notesv1.GetNoteResp, error) {
	return s.GetNoteFn(ctx, req)
}

func (s *FakeNotesServer) ListNote(
	ctx context.Context,
	req *notesv1.ListNoteReq,
) (*notesv1.ListNoteResp, error) {
	return s.ListNoteFn(ctx, req)
}

func (s *FakeNotesServer) UpdateNote(
	ctx context.Context,
	req *notesv1.UpdateNoteReq,
) (*notesv1.UpdateNoteResp, error) {
	return s.UpdateNoteFn(ctx, req)
}

func (s *FakeNotesServer) DeleteNote(
	ctx context.Context,
	req *notesv1.DeleteNoteReq,
) (*notesv1.DeleteNoteResp, error) {
	return s.DeleteNoteFn(ctx, req)
}
