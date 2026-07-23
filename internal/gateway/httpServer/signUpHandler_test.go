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

const fixtureSignUpHandler = `testdata\fixtures\signup\rows.sql`

func Test_signUpHandler(t *testing.T) {
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
				body: `{"login":"user","password":"11111111","email":"user@gmail.com"}`,
				mode: bodyAccessAndRefresh,
			},
			want: wantResp{
				code: http.StatusCreated,
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
			name: "#03_BAD_EMAIL",
			req: wantReq{
				body: `{"login":"user","password":"11111111","email":"usermail.com"}`,
				mode: bodyRaw,
			},
			want: wantResp{
				code: http.StatusBadRequest,
				body: `{"error":"bad request"}`,
			},
		},
		{
			name: "#04_SHORT_PASSWORD",
			req: wantReq{
				body: `{"login":"user","password":"11","email":"usermail.com"}`,
				mode: bodyRaw,
			},
			want: wantResp{
				code: http.StatusBadRequest,
				body: `{"error":"bad request"}`,
			},
		},
		{
			name: "05_ALREADY_EXISTS",
			req: wantReq{
				body: `{"login":"alesha","password":"11111111","email":"alesha@gmail.com"}`,
				mode: bodyRaw,
			},
			want: wantResp{
				code: http.StatusConflict,
				body: `{"error":"already exists"}`,
			},
		},
	}

	sut := middlewares.LogMiddleware(signUpHandler(svcAuth))
	method := http.MethodPost
	hndURL := "/auth/signup"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, testutil.LoadFixtures(pool, fixtureSignUpHandler, resetALLFixtures))
			req := httptest.NewRequest(method, hndURL, strings.NewReader(tt.req.body))
			rr := httptest.NewRecorder()
			sut.ServeHTTP(rr, req)

			body := parseResponse(t, rr, tt.req.mode)
			assert.Equal(t, tt.want.code, rr.Code)
			assert.Equal(t, testutil.NormalizeJSON(t, tt.want.body), testutil.NormalizeJSON(t, body))
		})
	}
}
