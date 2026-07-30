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

const fixtureSignInHandler = `testdata/fixtures/signin/rows.sql`

func Test_signInHandler(t *testing.T) {
	type wantReq struct {
		body string
		mode authBodyMode
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
				body: `{"login":"alesha","password":"11111111"}`,
				mode: bodyAccessAndRefresh,
			},
			want: wantResp{
				code: http.StatusOK,
				body: "OK",
			},
		},
		{
			name: "#02_BAD_JSON",
			req: wantReq{
				body: `login`,
				mode: bodyRaw,
			},
			want: wantResp{
				code: http.StatusUnprocessableEntity,
				body: `{"error":"invalid json"}`,
			},
		},
		{
			name: "#03_WRONG_PASSWORD",
			req: wantReq{
				body: `{"login":"alesha","password":"111"}`,
				mode: bodyRaw,
			},
			want: wantResp{
				code: http.StatusBadRequest,
				body: `{"error":"bad request"}`,
			},
		},
		{
			name: "#04_WRONG_USERNAME",
			req: wantReq{
				body: `{"login":"user","password":"11111111"}`,
				mode: bodyRaw,
			},
			want: wantResp{
				code: http.StatusUnauthorized,
				body: `{"error":"unauthorized"}`,
			},
		},
	}

	sut := middlewares.LogMiddleware(signInHandler(svcAuth))
	method := http.MethodPost
	hndURL := "/auth/signin"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, testutil.LoadFixtures(pool, fixtureSignInHandler, resetALLFixtures))
			req := httptest.NewRequest(method, hndURL, strings.NewReader(tt.req.body))
			rr := httptest.NewRecorder()
			sut.ServeHTTP(rr, req)

			body := parseResponse(t, rr, tt.req.mode)
			assert.Equal(t, tt.want.code, rr.Code)
			assert.Equal(t, testutil.NormalizeJSON(t, tt.want.body),
				testutil.NormalizeJSON(t, body))
		})
	}
}
