package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	userservice "notes/internal/gateway/service/note"
	"notes/internal/pkg/middlewares"
	"notes/internal/pkg/testutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_createNoteHandler(t *testing.T) {
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
				body:  `{"title":"test title","body":"test body"}`,
				ctxID: 101,
			},
			want: wantResp{
				code: http.StatusCreated,
				body: `{"id":1}`,
			},
		},
		{
			name: "#02_Invalid_json",
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
			name: "#03_Wrong_fields",
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
			name: "#04_Invalid_id",
			req: wantReq{
				body:  `{"title":"x","body":"y"}`,
				ctxID: 0,
			},
			want: wantResp{
				code: http.StatusBadRequest,
				body: `{"error":"bad request"}`,
			},
		},
		{
			name: "#05_Empty_title",
			req: wantReq{
				body:  `{"title":"","body":"y"}`,
				ctxID: 101,
			},
			want: wantResp{
				code: http.StatusBadRequest,
				body: `{"error":"bad request"}`,
			},
		},
		{
			name: "#06_NOT_FOUND",
			req: wantReq{
				body:  `{"title":"test title","body":"test body"}`,
				ctxID: 102,
			},
			want: wantResp{
				code: http.StatusNotFound,
				body: `{"error":"not found"}`,
			},
		},
	}

	repo := newTestRepo(t)
	svc := userservice.NewService(repo)

	sut := middlewares.LogMiddleware(createNoteHandler(svc))

	method := http.MethodPost
	hndURL := "/note/create"

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
}
