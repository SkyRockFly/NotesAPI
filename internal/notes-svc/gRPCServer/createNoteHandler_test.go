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
)

func TestServer_CreateNote(t *testing.T) {
	type wantReq struct {
		req *notesv1.CreateNoteReq
	}
	type wantResp struct {
		id   int32
		code codes.Code
	}
	tests := []struct {
		name string
		req  wantReq
		want wantResp
	}{
		{
			name: "#01_OK",
			req: wantReq{
				req: &notesv1.CreateNoteReq{
					AccountId: 101,
					Title:     "test title",
					Body:      "test body",
				},
			},
			want: wantResp{
				id:   5,
				code: codes.OK,
			},
		},
		{
			name: "#02_BAD_FIELDS",
			req: wantReq{
				req: &notesv1.CreateNoteReq{
					AccountId: 0,
					Title:     "x",
					Body:      "y",
				},
			},
			want: wantResp{
				code: codes.InvalidArgument,
			},
		},
		{
			name: "#03_INVALID_ID",
			req: wantReq{
				req: &notesv1.CreateNoteReq{
					AccountId: 101,
					Title:     "",
					Body:      "y",
				},
			},
			want: wantResp{
				code: codes.InvalidArgument,
			},
		},
		{
			name: "#04_INVALID_TITLE",
			req: wantReq{
				req: &notesv1.CreateNoteReq{
					AccountId: 101,
					Title:     "",
					Body:      "y",
				},
			},
			want: wantResp{
				code: codes.InvalidArgument,
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

			resp, err := client.CreateNote(ctx, tt.req.req)
			require.Equal(t, tt.want.code, status.Code(err))

			if tt.want.code != codes.OK {
				require.Error(t, err)
				require.Nil(t, resp)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Equal(t, tt.want.id, resp.GetId())
		})
	}
}
