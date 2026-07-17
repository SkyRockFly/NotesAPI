package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	notesvc "notes/internal/gateway/service/note"
	"notes/internal/pkg/middlewares"
	"notes/internal/pkg/testutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_updateNoteHandler(t *testing.T) {
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
				body:  `{"id":1,"title":"title update","body":"body update"}`,
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
				body:  `{"id":2,"title":"hello","body":"there"}`,
				ctxID: 101,
			},
			want: wantResp{
				code: http.StatusNotFound,
				body: `{"error":"not found"}`,
			},
		},
		{
			name: "#03_INVALID_JSON",
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
			name: "#04_WRONG_FIELDS",
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
			name: "#05_INVALID_ID",
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
			name: "#06_INVALID_ACCOUNT_ID",
			req: wantReq{
				body:  `{"id":2}`,
				ctxID: 0,
			},
			want: wantResp{
				code: http.StatusBadRequest,
				body: `{"error":"bad request"}`,
			},
		},
		{
			name: "#07_EMPTY_TITLE",
			req: wantReq{
				body:  `{"id":2,"title":""}`,
				ctxID: 101,
			},
			want: wantResp{
				code: http.StatusBadRequest,
				body: `{"error":"bad request"}`,
			},
		},
	}

	repo := newTestRepo(t)
	svc := notesvc.NewService(repo)

	sut := updateNoteHandler(svc)

	method := http.MethodPatch
	hndURL := "/note/update"

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
