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

func TestHTTPListNoteHandler(t *testing.T) {
	type wantReq struct {
		req *notesv1.ListNoteReq
	}

	type wantResp struct {
		code codes.Code
		resp *notesv1.ListNoteResp
	}

	tests := []struct {
		name string
		req  wantReq
		want wantResp
	}{
		{
			name: "#01_OK",
			req: wantReq{
				req: &notesv1.ListNoteReq{
					Cursor:    0,
					Limit:     10,
					Next:      true,
					AccountId: 101,
				},
			},
			want: wantResp{
				code: codes.OK,
				resp: &notesv1.ListNoteResp{
					CursorNext: 2,
					CursorPrev: 1,
					Notes: []*notesv1.Note{
						{
							Id:        1,
							AccountId: 101,
							Title:     "get-visible-note",
							Body:      "Fixture used by the GET note test.",
							CreatedAt: timestamppb.New(
								time.Date(
									2025,
									time.August,
									12,
									8,
									0,
									0,
									0,
									time.UTC,
								),
							),
							UpdatedAt: timestamppb.New(
								time.Date(
									2025,
									time.August,
									12,
									8,
									15,
									0,
									0,
									time.UTC,
								),
							),
						},
						{
							Id:        2,
							AccountId: 101,
							Title:     "another note",
							Body:      "for list",
							CreatedAt: timestamppb.New(
								time.Date(
									2025,
									time.August,
									12,
									9,
									30,
									0,
									0,
									time.UTC,
								),
							),
							UpdatedAt: timestamppb.New(
								time.Date(
									2025,
									time.August,
									12,
									9,
									45,
									0,
									0,
									time.UTC,
								),
							),
						},
					},
					HasMore: false,
				},
			},
		},
		{
			name: "#02_OK_REVERSE",
			req: wantReq{
				req: &notesv1.ListNoteReq{
					Cursor:    2,
					Limit:     10,
					Next:      false,
					AccountId: 101,
				},
			},
			want: wantResp{
				code: codes.OK,
				resp: &notesv1.ListNoteResp{
					CursorNext: 1,
					CursorPrev: 1,
					Notes: []*notesv1.Note{
						{
							Id:        1,
							AccountId: 101,
							Title:     "get-visible-note",
							Body:      "Fixture used by the GET note test.",
							CreatedAt: timestamppb.New(
								time.Date(
									2025,
									time.August,
									12,
									8,
									0,
									0,
									0,
									time.UTC,
								),
							),
							UpdatedAt: timestamppb.New(
								time.Date(
									2025,
									time.August,
									12,
									8,
									15,
									0,
									0,
									time.UTC,
								),
							),
						},
					},
					HasMore: false,
				},
			},
		},
		{
			name: "#03_INVALID_ID",
			req: wantReq{
				req: &notesv1.ListNoteReq{
					Cursor:    0,
					Limit:     0,
					Next:      false,
					AccountId: 0,
				},
			},
			want: wantResp{
				code: codes.InvalidArgument,
			},
		},
		{
			name: "#04_REVERSE_FROM_ZERO",
			req: wantReq{
				req: &notesv1.ListNoteReq{
					Cursor:    0,
					Limit:     10,
					Next:      false,
					AccountId: 101,
				},
			},
			want: wantResp{
				code: codes.InvalidArgument,
			},
		},
		{
			name: "#05_GET_NOT_EXISTING_NOTES",
			req: wantReq{
				req: &notesv1.ListNoteReq{
					AccountId: 1010101,
					Limit:     10,
					Cursor:    0,
					Next:      true,
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
			resp, err := client.ListNote(ctx, tt.req.req)

			require.Equal(t, tt.want.code, status.Code(err))

			if tt.want.code != codes.OK {
				require.Error(t, err)
				require.Nil(t, resp)
				return
			}

			require.NoError(t, err)
			require.True(t, proto.Equal(tt.want.resp, resp))
		})
	}
}
