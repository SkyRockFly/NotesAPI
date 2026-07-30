package grpcserver

import (
	"context"
	notesv1 "notes/internal/grpc/generated"
	noteservice "notes/internal/notes-svc/noteService"
)

func (s *Server) DeleteNote(ctx context.Context, req *notesv1.DeleteNoteReq) (*notesv1.DeleteNoteResp, error) {
	logger := getCtxLogger(ctx)
	svcReq := noteservice.DeleteReq{
		ID:        int(req.GetId()),
		AccountID: int(req.GetAccountId()),
	}
	if err := s.service.Delete(ctx, svcReq); err != nil {
		return nil, handleGRPCError(err, logger)
	}

	return &notesv1.DeleteNoteResp{}, nil
}
