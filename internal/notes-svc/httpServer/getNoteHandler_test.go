package httpserver

import (
	"net/http"
	"net/http/httptest"
	"notes/internal/pkg/testutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPGetNoteHandler(t *testing.T) {
	require.NoError(t, testutil.LoadFixtures(pool, fixturePath, resetFixtures))
	sut := HTTPGetNoteHandler(svc)

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
				body: `{"account_id":101,"id":1}`,
			},
			want: wantResp{
				code: http.StatusOK,
				body: `{"id":1,"account_id":101,"title":"get-visible-note",
"body":"Fixture used by the GET note test.",
"created_at":"2025-08-12T08:00:00Z","updated_at":"2025-08-12T08:15:00Z"}`,
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
			name: "#04_INVALID_IDS",
			req: wantReq{
				body: `{"account_id":0,"id":0}`,
			},
			want: wantResp{
				code: http.StatusBadRequest,
				body: `{"error":"bad request"}`,
			},
		},
		{
			name: "#05_GET_NOT_EXISTING",
			req: wantReq{
				body: `{"account_id":101,"id":99999}`,
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
