package grpcserver

import (
	"context"
	notesv1 "notes/internal/grpc/generated"
	"notes/internal/pkg/testutil"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestHTTPGetNoteHandler(t *testing.T) {
	type wantReq struct {
		req *notesv1.GetNoteReq
	}
	type wantResp struct {
		code codes.Code
		note *notesv1.GetNoteResp
	}

	tests := []struct {
		name string
		req  wantReq
		want wantResp
	}{
		{
			name: "#01_OK",
			req: wantReq{
				req: &notesv1.GetNoteReq{
					AccountId: 101,
					Id:        1,
				},
			},
			want: wantResp{
				code: codes.OK,
				note: &notesv1.GetNoteResp{
					Id:        1,
					AccountId: 101,
					Title:     "get-visible-note",
					Body:      "Fixture used by the GET note test.",
					CreatedAt: timestamppb.New(
						time.Date(2025, time.August, 12, 8, 0, 0, 0, time.UTC),
					),
					UpdatedAt: timestamppb.New(
						time.Date(2025, time.August, 12, 8, 15, 0, 0, time.UTC),
					),
				},
			},
		},
		{
			name: "#02_INVALID_IDS",
			req: wantReq{
				req: &notesv1.GetNoteReq{
					AccountId: 0,
					Id:        0,
				},
			},
			want: wantResp{
				code: codes.InvalidArgument,
			},
		},
		{
			name: "#03_GET_NOT_EXISTING",
			req: wantReq{
				req: &notesv1.GetNoteReq{
					AccountId: 101,
					Id:        99999,
				},
			},
			want: wantResp{
				code: codes.NotFound,
			},
		},
	}

	client := newTestGRPCClient(t, handler)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, testutil.LoadFixtures(pool, fixturePath, resetFixtures))

			ctx, cancel := context.WithTimeout(
				context.Background(),
				3*time.Second,
			)
			defer cancel()

			resp, err := client.GetNote(ctx, tt.req.req)
			require.Equal(t, tt.want.code, status.Code(err))

			if tt.want.code != codes.OK {
				require.Error(t, err)
				require.Nil(t, resp)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.True(t, proto.Equal(tt.want.note, resp))
		})
	}
}
