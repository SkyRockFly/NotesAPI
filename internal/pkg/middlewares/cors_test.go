package middlewares

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCORS(t *testing.T) {
	type wantReq struct {
		origin        string
		options       bool
		allowedOrigin bool
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
				origin:        "http://localhost:5682",
				allowedOrigin: true,
			},
			want: wantResp{
				code: http.StatusOK,
				body: "passed",
			},
		},
		{
			name: "#02_BAD_ORIGIN",
			req: wantReq{
				origin:        "lmao:5682",
				allowedOrigin: false,
			},
			want: wantResp{
				code: http.StatusOK,
				body: "passed",
			},
		},
		{
			name: "#03_OPTIONS",
			req: wantReq{
				origin:        "http://localhost:5682",
				options:       true,
				allowedOrigin: true,
			},
			want: wantResp{
				code: http.StatusNoContent,
				body: ``,
			},
		},
		{
			name: "#03_OPTIONS_WITH_BAD_ORIGIN",
			req: wantReq{
				origin:        "http://lmao:5682",
				options:       true,
				allowedOrigin: false,
			},
			want: wantResp{
				code: http.StatusForbidden,
				body: `{"error":"origin is not allowed"}`,
			},
		},
	}

	const allowedOrigin = "http://localhost:5682"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("passed"))
			})
			handler := CORS(next, allowedOrigin)

			method := http.MethodPost
			if tt.req.options {
				method = http.MethodOptions
			}
			req := httptest.NewRequest(method, "/", strings.NewReader(`{}`))
			req.Header.Add("Origin", tt.req.origin)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.want.code, rec.Code)
			assert.Equal(t, tt.want.body, rec.Body.String())

			expected := "http://localhost:5682"
			if !tt.req.allowedOrigin {
				expected = ""
			}
			assert.Equal(
				t,
				expected,
				rec.Header().Get("Access-Control-Allow-Origin"),
			)
		})
	}
}
