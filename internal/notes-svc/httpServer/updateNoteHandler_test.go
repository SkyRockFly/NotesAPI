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

func TestHTTPUpdateNoteHandler(t *testing.T) {
	sut := middlewares.LogMiddleware(HTTPUpdateNoteHandler(svc))

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
				body: `{"account_id":202,"id":3,"title":"brr"}`,
			},
			want: wantResp{
				code: http.StatusNoContent,
				body: "",
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
			name: "#04_INVALID_ID",
			req: wantReq{
				body: `{"account_id":0,"title":"x","body":"y","id":0}`,
			},
			want: wantResp{
				code: http.StatusBadRequest,
				body: `{"error":"bad request"}`,
			},
		},

		{
			name: "#05_UPDATE_NON_EXISTING",
			req: wantReq{
				body: `{"account_id":202,"id":99999,"title":"brr"}`,
			},
			want: wantResp{
				code: http.StatusNotFound,
				body: `{"error":"not found"}`,
			},
		},
	}

	url := "/note/update"
	method := http.MethodPost

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, testutil.LoadFixtures(pool, fixturePath, resetFixtures))
			req := httptest.NewRequest(method, url, strings.NewReader(tt.req.body))
			rr := httptest.NewRecorder()

			sut.ServeHTTP(rr, req)
			assert.Equal(t, tt.want.code, rr.Code, "status code")
			assert.Equal(t, testutil.NormalizeJSON(t, tt.want.body), testutil.NormalizeJSON(t, rr.Body.String()))
		})
	}
}
