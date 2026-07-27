package httpserver

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"notes/internal/pkg/apperror"
	"notes/internal/pkg/testutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const logoutFixtures = `testdata/fixtures/logout/rows.sql`

func TestLogoutHandler(t *testing.T) {
	type wantReq struct {
		refresh string
		svcErr  error
	}

	type wantResp struct {
		code          int
		cookieDeleted bool
		isRevoked     bool
	}

	tests := []struct {
		name string
		req  wantReq
		want wantResp
	}{
		{
			name: "#01_OK",
			req: wantReq{
				refresh: "18f11757-c9cf-47f4-a187-ddfda409abb4.JZO-pIBrPciCvkUdarnkUWzitwphl3rU5P0xFDPHEo4",
			},
			want: wantResp{
				code:          http.StatusNoContent,
				cookieDeleted: true,
				isRevoked:     true,
			},
		},
		{
			name: "#02_NO_COOKIE",
			req:  wantReq{},
			want: wantResp{
				code:          http.StatusNoContent,
				cookieDeleted: true,
				isRevoked:     false,
			},
		},
		{
			name: "#03_NOT_FOUND",
			req: wantReq{
				refresh: "selector.private",
			},
			want: wantResp{
				code:          http.StatusNoContent,
				cookieDeleted: true,
				isRevoked:     false,
			},
		},
	}

	handler := logoutHandler(svcAuth)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, testutil.LoadFixtures(pool, logoutFixtures, resetALLFixtures))

			req := httptest.NewRequest(
				http.MethodPost,
				"/auth/logout",
				nil,
			)

			if tt.req.refresh != "" {
				req.AddCookie(&http.Cookie{
					Name:  "refresh_token",
					Value: tt.req.refresh,
					Path:  "/auth",
				})
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.want.code, rec.Code)
			assert.Equal(
				t,
				tt.want.cookieDeleted,
				refreshCookieDeleted(rec),
			)
			if tt.want.isRevoked {
				assert.True(t, checkRefreshRevoke(t, tt.req.refresh))
			}
		})
	}
}

func checkRefreshRevoke(t *testing.T, refresh string) bool {
	t.Helper()
	selector, _, err := getRefresh(refresh)
	require.NoError(t, err)

	sql := `SELECT revoked FROM refresh_token WHERE selector = $1`
	var revoked bool
	err = pool.QueryRow(context.Background(), sql, selector).Scan(&revoked)
	require.NoError(t, err)

	return revoked
}

func refreshCookieDeleted(rec *httptest.ResponseRecorder) bool {
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name != "refresh_token" {
			continue
		}

		return cookie.Value == "" &&
			cookie.Path == "/auth" &&
			cookie.MaxAge < 0
	}

	return false
}

func getRefresh(refresh string) (string, []byte, error) {
	parts := strings.Split(refresh, ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", nil, fmt.Errorf("%w: invalid refresh format", apperror.ErrBadRequest)
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", nil, fmt.Errorf("%w: invalid private part", apperror.ErrBadRequest)
	}

	return parts[0], raw, nil
}
