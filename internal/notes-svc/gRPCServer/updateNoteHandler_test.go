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

func TestHTTPUpdateNoteHandler(t *testing.T) {
	type wantReq struct {
		req *notesv1.UpdateNoteReq
	}

	type wantResp struct {
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
				req: &notesv1.UpdateNoteReq{
					AccountId: 202,
					Id:        3,
					Title:     "brr",
				},
			},
			want: wantResp{
				code: codes.OK,
			},
		},
		{
			name: "#02_INVALID_ID",
			req: wantReq{
				req: &notesv1.UpdateNoteReq{
					AccountId: 0,
					Id:        0,
					Title:     "x",
					Body:      "y",
				},
			},
			want: wantResp{
				code: codes.InvalidArgument,
			},
		},
		{
			name: "#03_UPDATE_NON_EXISTING",
			req: wantReq{
				req: &notesv1.UpdateNoteReq{
					AccountId: 202,
					Id:        99999,
					Title:     "brr",
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
			resp, err := client.UpdateNote(ctx, tt.req.req)

			require.Equal(t, tt.want.code, status.Code(err))

			if tt.want.code != codes.OK {
				require.Error(t, err)
				require.Nil(t, resp)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
		})
	}
}
