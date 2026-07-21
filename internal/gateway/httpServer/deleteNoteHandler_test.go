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

func Test_deleteNoteHandler(t *testing.T) {
	type wantReq struct {
		body  string
		ctxID int
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
				body:  `{"id":1}`,
				ctxID: 101,
			},
			want: wantResp{
				code: http.StatusNoContent,
				body: "",
			},
		},
		{
			name: "#02_NON_EXISTING_NOTE",
			req: wantReq{
				body:  `{"id":2}`,
				ctxID: 101,
			},
			want: wantResp{
				code: http.StatusNotFound,
				body: `{"error":"not found"}`,
			},
		},
		{
			name: "#03_Invalid_json",
			req: wantReq{
				body:  `{"account_id":101,`,
				ctxID: 101,
			},
			want: wantResp{
				code: http.StatusUnprocessableEntity,
				body: `{"error":"invalid json"}`,
			},
		},
		{
			name: "#04_Wrong_fields",
			req: wantReq{
				body:  `{"acc_id":101,"ttl":"x"}`,
				ctxID: 101,
			},
			want: wantResp{
				code: http.StatusUnprocessableEntity,
				body: `{"error":"invalid json"}`,
			},
		},
		{
			name: "#05_Invalid_id",
			req: wantReq{
				body:  `{"id":0}`,
				ctxID: 101,
			},
			want: wantResp{
				code: http.StatusBadRequest,
				body: `{"error":"bad request"}`,
			},
		},
		{
			name: "#06_Invalid_AccountID",
			req: wantReq{
				body:  `{"id":2}`,
				ctxID: 0,
			},
			want: wantResp{
				code: http.StatusBadRequest,
				body: `{"error":"bad request"}`,
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

	method := http.MethodDelete
	hndURL := "/note/delete"

	for _, repo := range repos {
		t.Run(repo.name, func(t *testing.T) {
			svc := userservice.NewService(repo.repo)
			sut := middlewares.LogMiddleware(deleteNoteHandler(svc))
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
