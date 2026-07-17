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

func Test_listNoteHandler(t *testing.T) {
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
				body:  `{"cursor":0,"limit":10,"next":true}`,
				ctxID: 101,
			},
			want: wantResp{
				code: http.StatusOK,
				body: `{"cursor_next":2,"cursor_prev":1,"notes":[{"id":1,"account_id":101,"title":"test title","body":"test body",
"created_at":"2026-07-13T20:00:00Z","updated_at":"2026-07-13T20:00:00Z"},
{"id":2,"account_id":101,"title":"test title","body":"test body","created_at":
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
	}

	repo := newTestRepo(t)
	svc := notesvc.NewService(repo)

	sut := middlewares.LogMiddleware(listNoteHandler(svc))

	method := http.MethodPost
	hndURL := "/notes/get"

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
