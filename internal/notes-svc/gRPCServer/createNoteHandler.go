package grpcserver

import (
	"context"
	notesv1 "notes/internal/grpc/generated"
	noteservice "notes/internal/notes-svc/noteService"
)

func (s *Server) CreateNote(ctx context.Context, req *notesv1.CreateNoteReq) (*notesv1.CreateNoteResp, error) {
	logger := getCtxLogger(ctx)
	svcReq := noteservice.CreateReq{
		AccountID: int(req.GetAccountId()),
		Title:     req.GetTitle(),
		Body:      req.GetBody(),
	}
	id, err := s.service.Create(ctx, svcReq)
	if err != nil {
		return nil, handleGRPCError(err, logger)
	}

	resp := &notesv1.CreateNoteResp{
		Id: int32(id),
	}

	return resp, nil
}
