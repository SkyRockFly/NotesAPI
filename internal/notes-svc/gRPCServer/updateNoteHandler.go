package grpcserver

import (
	"context"
	notesv1 "notes/internal/grpc/generated"
	noteservice "notes/internal/notes-svc/noteService"
)

func (s *Server) UpdateNote(ctx context.Context, req *notesv1.UpdateNoteReq) (*notesv1.UpdateNoteResp, error) {
	logger := getCtxLogger(ctx)
	svcReq := noteservice.UpdateReq{
		ID:        int(req.GetId()),
		AccountID: int(req.GetAccountId()),
		Title:     req.GetTitle(),
		Body:      req.GetBody(),
	}
	if err := s.service.Update(ctx, svcReq); err != nil {
		return nil, handleGRPCError(err, logger)
	}

	return &notesv1.UpdateNoteResp{}, nil
}
