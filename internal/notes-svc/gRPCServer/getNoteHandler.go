package grpcserver

import (
	"context"
	notesv1 "notes/internal/grpc/generated"
	noteservice "notes/internal/notes-svc/noteService"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) GetNote(ctx context.Context, req *notesv1.GetNoteReq) (*notesv1.GetNoteResp, error) {
	logger := getCtxLogger(ctx)
	svcReq := noteservice.GetReq{
		ID:        int(req.GetId()),
		AccountID: int(req.GetAccountId()),
	}
	note, err := s.service.Get(ctx, svcReq)
	if err != nil {
		return nil, handleGRPCError(err, logger)
	}

	resp := &notesv1.GetNoteResp{
		Id:        int32(note.ID),
		AccountId: int32(note.AccountID),
		Title:     note.Title,
		Body:      note.Body,
		CreatedAt: timestamppb.New(note.CreatedAt),
		UpdatedAt: timestamppb.New(note.UpdatedAt),
	}

	return resp, nil
}
