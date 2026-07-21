package grpcserver

import (
	"context"
	notesv1 "notes/internal/grpc/generated"
	noteservice "notes/internal/notes-svc/noteService"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) ListNote(ctx context.Context, req *notesv1.ListNoteReq) (*notesv1.ListNoteResp, error) {
	logger := getCtxLogger(ctx)
	svcReq := noteservice.ListReq{
		AccountID: int(req.GetAccountId()),
		Limit:     int(req.GetLimit()),
		Cursor:    int(req.GetCursor()),
		Next:      req.GetNext(),
	}
	svcResp, err := s.service.List(ctx, svcReq)
	if err != nil {
		return nil, handleGRPCError(err, logger)
	}

	resp := &notesv1.ListNoteResp{
		CursorNext: int32(svcResp.CursorNext),
		CursorPrev: int32(svcResp.CursorPrev),
		HasMore:    svcResp.HasMore,
		Notes:      mapNotesToGRPCNotes(svcResp.Notes),
	}

	return resp, nil
}

func mapNotesToGRPCNotes(svc []noteservice.Note) []*notesv1.Note {
	grpcNotes := make([]*notesv1.Note, 0, len(svc))

	for _, note := range svc {
		grpcNote := &notesv1.Note{
			Id:        int32(note.ID),
			AccountId: int32(note.AccountID),
			Title:     note.Title,
			Body:      note.Body,
			CreatedAt: timestamppb.New(note.CreatedAt),
			UpdatedAt: timestamppb.New(note.UpdatedAt),
		}
		grpcNotes = append(grpcNotes, grpcNote)
	}

	return grpcNotes
}
