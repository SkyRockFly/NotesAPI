package httpserver

import (
	"net/http"
	"net/http/httptest"
	"notes/internal/pkg/middlewares"
	"notes/internal/pkg/testutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPListNoteHandler(t *testing.T) {
	require.NoError(t, testutil.LoadFixtures(pool, fixturePath, resetFixtures))
	sut := middlewares.LogMiddleware(HTTPListNoteHandler(svc))

	type wantReq struct {
		body string
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
				body: `{"cursor":0,"limit":10,"next":true,"account_id":101}`,
			},
			want: wantResp{
				code: http.StatusOK,
				body: `{"cursor_next":2,"cursor_prev":1,"notes":
[{"id":1,"account_id":101,"title":"get-visible-note","body":"Fixture used by the GET note test.",
"created_at":"2025-08-12T08:00:00Z","updated_at":"2025-08-12T08:15:00Z"},
{"id":2,"account_id":101,"title":"another note","body":"for list","created_at":"2025-08-12T09:30:00Z","updated_at":"2025-08-12T09:45:00Z"}]
,"has_more":false}`,
			},
		},
		{
			name: "#02_OK_REVERSE",
			req: wantReq{
				body: `{"cursor":2,"limit":10,"next":false,"account_id":101}`,
			},
			want: wantResp{
				code: http.StatusOK,
				body: `{"cursor_next":1,"cursor_prev":1,"notes":
[{"id":1,"account_id":101,"title":"get-visible-note","body":"Fixture used by the GET note test.","created_at":"2025-08-12T08:00:00Z",
"updated_at":"2025-08-12T08:15:00Z"}],"has_more":false}`,
			},
		},
		{
			name: "#03_INVALID_JSON",
			req: wantReq{
				body: `{"tit":101,`,
			},
			want: wantResp{
				code: http.StatusUnprocessableEntity,
				body: `{"error":"invalid json"}`,
			},
		},
		{
			name: "#04_WRONG_FIELDS",
			req: wantReq{
				body: `{"ttl":"x"}`,
			},
			want: wantResp{
				code: http.StatusUnprocessableEntity,
				body: `{"error":"invalid json"}`,
			},
		},
		{
			name: "#05_INVALID_ID",
			req: wantReq{
				body: `{"cursor":0,"limit":0,"next":false}`,
			},
			want: wantResp{
				code: http.StatusBadRequest,
				body: `{"error":"bad request"}`,
			},
		},

		{
			name: "#06_REVERSE_FROM_ZERO",
			req: wantReq{
				body: `{"cursor":0,"limit":0,"next":false}`,
			},
			want: wantResp{
				code: http.StatusBadRequest,
				body: `{"error":"bad request"}`,
			},
		},

		{
			name: "#07_GET_NOT_EXISTING_NOTES",
			req: wantReq{
				body: `{"account_id":1010101,"limit":10,"cursor":0,"next":true}`,
			},
			want: wantResp{
				code: http.StatusNotFound,
				body: `{"error":"not found"}`,
			},
		},
	}

	url := "/note/create"
	method := http.MethodPost

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(method, url, strings.NewReader(tt.req.body))
			rr := httptest.NewRecorder()

			sut.ServeHTTP(rr, req)
			assert.Equal(t, tt.want.code, rr.Code, "status code")
			assert.Equal(t, testutil.NormalizeJSON(t, tt.want.body), testutil.NormalizeJSON(t, rr.Body.String()))
		})
	}
}
