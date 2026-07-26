package middlewares

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuth(t *testing.T) {
	type wantReq struct {
		subject       string
		withoutHeader bool
		customJWT     string
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
				subject: "10",
			},
			want: wantResp{
				code: http.StatusOK,
				body: "passed",
			},
		},
		{
			name: "#02_WITHOUT_HEADER",
			req: wantReq{
				withoutHeader: true,
			},
			want: wantResp{
				code: http.StatusUnauthorized,
				body: `{"error":"invalid auth header"}`,
			},
		},
		{
			name: "#03_BAD_JWT",
			req: wantReq{
				customJWT: "lmao",
			},
			want: wantResp{
				code: http.StatusUnauthorized,
				body: `{"error":"unauthorized"}`,
			},
		},
		{
			name: "#04_BAD_SUBJECT",
			req: wantReq{
				subject: "lmao",
			},
			want: wantResp{
				code: http.StatusInternalServerError,
				body: `{"error":"service error"}`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()
				uid, ok := ctx.Value(UIDKey).(int)
				require.True(t, ok, "extract value")
				require.Greater(t, uid, 0)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("passed"))
			})
			auth := Auth([]byte("le secret"), 5*time.Second)
			handler := auth(next)

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
			rec := httptest.NewRecorder()

			if !tt.req.withoutHeader {
				token := tt.req.customJWT
				if token == "" {
					var err error
					token, err = generateJWTToken([]byte("le secret"), tt.req.subject)
					require.NoError(t, err)
				}
				req.Header.Set("Authorization", "Bearer "+token)
			}

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.want.code, rec.Code)
			assert.Equal(t, tt.want.body, rec.Body.String())
		})
	}
}

func generateJWTToken(secret []byte, subject string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   subject,
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Minute * 10)),
	})

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("signing jwt: %w", err)
	}

	return tokenString, nil
}
