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

func TestHTTPCreateNoteHandler(t *testing.T) {
	require.NoError(t, testutil.LoadFixtures(pool, fixturePath, resetFixtures))
	sut := middlewares.LogMiddleware(HTTPCreateNoteHandler(svc))

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
				body: `{"account_id":101,"title":"test title","body":"test body"}`,
			},
			want: wantResp{
				code: http.StatusCreated,
				body: `{"id":6}`,
			},
		},
		{
			name: "#02_INVALID_JSON",
			req: wantReq{
				body: `{"account_id":101,`,
			},
			want: wantResp{
				code: http.StatusUnprocessableEntity,
				body: `{"error":"invalid json"}`,
			},
		},
		{
			name: "#03_BAD_FIELDS",
			req: wantReq{
				body: `{"acc_id":101,"ttl":"x"}`,
			},
			want: wantResp{
				code: http.StatusUnprocessableEntity,
				body: `{"error":"invalid json"}`,
			},
		},
		{
			name: "04_INVALID_ID",
			req: wantReq{
				body: `{"account_id":0,"title":"x","body":"y"}`,
			},
			want: wantResp{
				code: http.StatusBadRequest,
				body: `{"error":"invalid content of fields"}`,
			},
		},
		{
			name: "05_INVALID_TITLE",
			req: wantReq{
				body: `{"account_id":1,"title":"","body":"y"}`,
			},
			want: wantResp{
				code: http.StatusBadRequest,
				body: `{"error":"invalid content of fields"}`,
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
			assert.Equal(t, testutil.NormalizeJSON(t, tt.want.body), testutil.NormalizeJSON(t, tt.want.body))
		})
	}
}
