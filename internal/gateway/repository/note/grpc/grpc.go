package notegrpc

import (
	"context"
	"fmt"
	noterepo "notes/internal/gateway/repository/note"
	notesv1 "notes/internal/grpc/generated"
	"notes/internal/pkg/apperror"
	requestid "notes/internal/pkg/requestID"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type GRPCRepo struct {
	client notesv1.NotesServiceClient
}

type SvcHTTPCfg struct {
	Host string
	Port int
}

func (r *GRPCRepo) Create(ctx context.Context, req noterepo.CreateReq) (int, error) {
	grpcReq := &notesv1.CreateNoteReq{
		AccountId: int32(req.AccountID),
		Title:     req.Title,
		Body:      req.Body,
	}

	ctx = makeRequestID(ctx)
	resp, err := r.client.CreateNote(ctx, grpcReq)
	if err != nil {
		st := status.Convert(err)

		log.Ctx(ctx).
			Error().
			Err(err).
			Str("grpc_code", st.Code().String()).
			Str("grpc_message", st.Message()).
			Msg("notes grpc call failed")

		return 0, fmt.Errorf("repo.Create:%w", mapGRPCError(status.Code(err)))
	}

	return int(resp.GetId()), nil
}

func (r *GRPCRepo) Delete(ctx context.Context, req noterepo.DeleteReq) error {
	grpcReq := &notesv1.DeleteNoteReq{
		Id:        int32(req.ID),
		AccountId: int32(req.AccountID),
	}
	_, err := r.client.DeleteNote(ctx, grpcReq)
	if err != nil {
		return fmt.Errorf("repo.Delete:%w", mapGRPCError(status.Code(err)))
	}

	return nil
}

func (r *GRPCRepo) Get(ctx context.Context, req noterepo.GetReq) (noterepo.Note, error) {
	grpcReq := &notesv1.GetNoteReq{
		Id:        int32(req.ID),
		AccountId: int32(req.AccountID),
	}

	ctx = makeRequestID(ctx)
	grpcResp, err := r.client.GetNote(ctx, grpcReq)
	if err != nil {
		return noterepo.Note{}, fmt.Errorf("repo.Get:%w", mapGRPCError(status.Code(err)))
	}

	resp := noterepo.Note{
		ID:        int(grpcResp.GetId()),
		AccountID: int(grpcResp.GetAccountId()),
		Title:     grpcResp.GetTitle(),
		Body:      grpcResp.GetBody(),
		CreatedAt: grpcResp.GetCreatedAt().AsTime(),
		UpdatedAt: grpcResp.GetUpdatedAt().AsTime(),
	}

	return resp, nil
}

func (r *GRPCRepo) List(ctx context.Context, req noterepo.ListReq) (noterepo.ListResp, error) {
	grpcReq := &notesv1.ListNoteReq{
		AccountId: int32(req.AccountID),
		Limit:     int32(req.Limit),
		Cursor:    int32(req.Cursor),
		Next:      req.Next,
	}

	ctx = makeRequestID(ctx)
	grpcResp, err := r.client.ListNote(ctx, grpcReq)
	if err != nil {
		return noterepo.ListResp{}, fmt.Errorf("repo.List:%w", mapGRPCError(status.Code(err)))
	}

	resp := noterepo.ListResp{
		CursorNext: int(grpcResp.GetCursorNext()),
		CursorPrev: int(grpcResp.GetCursorPrev()),
		HasMore:    grpcResp.GetHasMore(),
		Notes:      mapGRPCNoteToNote(grpcResp.GetNotes()),
	}

	return resp, nil
}

func (r *GRPCRepo) Update(ctx context.Context, req noterepo.UpdateReq) error {
	grpcReq := &notesv1.UpdateNoteReq{
		Id:        int32(req.ID),
		AccountId: int32(req.AccountID),
		Title:     req.Title,
		Body:      req.Body,
	}

	ctx = makeRequestID(ctx)
	if _, err := r.client.UpdateNote(ctx, grpcReq); err != nil {
		return fmt.Errorf("repo.Update:%w", mapGRPCError(status.Code(err)))
	}

	return nil
}

func NewRepo(client notesv1.NotesServiceClient) *GRPCRepo {
	return &GRPCRepo{
		client: client,
	}
}

func mapGRPCNoteToNote(grpcnotes []*notesv1.Note) []noterepo.Note {
	repoNotes := make([]noterepo.Note, 0, len(grpcnotes))
	for _, note := range grpcnotes {
		repoNote := noterepo.Note{
			ID:        int(note.GetId()),
			AccountID: int(note.GetAccountId()),
			Title:     note.GetTitle(),
			Body:      note.GetBody(),
			CreatedAt: note.GetCreatedAt().AsTime(),
			UpdatedAt: note.GetUpdatedAt().AsTime(),
		}
		repoNotes = append(repoNotes, repoNote)
	}
	return repoNotes
}

func makeRequestID(ctx context.Context) context.Context {
	id := requestid.FromContext(ctx)
	if id == "" {
		return ctx
	}

	return metadata.AppendToOutgoingContext(
		ctx,
		requestid.MetadataKey,
		id,
	)
}

func mapGRPCError(code codes.Code) error {
	switch code {
	case codes.OK:
		return nil

	case codes.InvalidArgument:
		return apperror.ErrBadRequest

	case codes.NotFound:
		return apperror.ErrNotFound

	default:
		return apperror.ErrBackend
	}
}
