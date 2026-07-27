package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	noterepo "notes/internal/gateway/repository/note"
	notegrpc "notes/internal/gateway/repository/note/grpc"
	userservice "notes/internal/gateway/service/note"
	"notes/internal/pkg/middlewares"
	"notes/internal/pkg/testutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_listNoteHandler(t *testing.T) {
	type wantReq struct {
		body  string
		ctxID any
	}
	type wantResp struct {
		code int
		body string
	}

	tests := []struct {
		name string
		req  wantReq
		want wantResp
	}{
		{
			name: "#01_OK",
			req: wantReq{
				body:  `{"cursor":0,"limit":10,"next":true}`,
				ctxID: 101,
			},
			want: wantResp{
				code: http.StatusOK,
				body: `{"cursor_next":2,"cursor_prev":1,"notes":[{"id":1,"title":"test title","body":"test body",
"created_at":"2026-07-13T20:00:00Z","updated_at":"2026-07-13T20:00:00Z"},
{"id":2,"title":"test title","body":"test body","created_at":
"2026-07-13T20:00:00Z","updated_at":"2026-07-13T20:00:00Z"}],"has_more":false}`,
			},
		},
		{
			name: "#02_Non-existing_accountID",
			req: wantReq{
				body:  `{"cursor":0,"limit":10,"next":true}`,
				ctxID: 202,
			},
			want: wantResp{
				code: http.StatusNotFound,
				body: `{"error":"not found"}`,
			},
		},
		{
			name: "#03_Invalid_AccountID",
			req: wantReq{
				body:  `{"cursor":0,"limit":10,"next":true}`,
				ctxID: 0,
			},
			want: wantResp{
				code: http.StatusBadRequest,
				body: `{"error":"bad request"}`,
			},
		},
		{
			name: "#04_NOT_EXISTING_ACCOUNT",
			req: wantReq{
				body:  `{"cursor":0,"limit":10,"next":true}`,
				ctxID: 202,
			},
			want: wantResp{
				code: http.StatusNotFound,
				body: `{"error":"not found"}`,
			},
		},
		{
			name: "#04_BAD_UID",
			req: wantReq{
				body:  `{"cursor":0,"limit":10,"next":true}`,
				ctxID: "lmao",
			},
			want: wantResp{
				code: http.StatusInternalServerError,
				body: `{"error":"service error"}`,
			},
		},
		{
			name: "#05_BAD_JSON",
			req: wantReq{
				body:  `{"cursor":0,"limit":10,"ne`,
				ctxID: 101,
			},
			want: wantResp{
				code: http.StatusUnprocessableEntity,
				body: `{"error":"invalid json"}`,
			},
		},
	}

	repos := []struct {
		name string
		repo noterepo.INote
	}{
		{
			name: "http",
			repo: newTestRepo(t),
		},
	}

	address := newTestGRPCServer(t)
	client, conn, err := makeGRPCClient(address)
	require.NoError(t, err)
	t.Cleanup(func() {
		conn.Close()
	})

	repos = append(repos, struct {
		name string
		repo noterepo.INote
	}{
		name: "grpc",
		repo: notegrpc.NewRepo(client),
	})

	method := http.MethodPost
	hndURL := "/notes/get"

	for _, repo := range repos {
		t.Run(repo.name, func(t *testing.T) {
			svc := userservice.NewService(repo.repo)
			sut := middlewares.LogMiddleware(listNoteHandler(svc))
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					req := httptest.NewRequest(method, hndURL, strings.NewReader(tt.req.body))

					rr := httptest.NewRecorder()
					idCtx := context.WithValue(req.Context(), middlewares.UIDKey, tt.req.ctxID)

					sut.ServeHTTP(rr, req.WithContext(idCtx))

					assert.Equal(t, tt.want.code, rr.Code)
					assert.Equal(t, testutil.NormalizeJSON(t, tt.want.body),
						testutil.NormalizeJSON(t, rr.Body.String()))
				})
			}
		})
	}
}
