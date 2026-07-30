package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"notes/internal/pkg/middlewares"
	"notes/internal/pkg/testutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const logoutAllFixtures = `testdata/fixtures/logoutAll/rows.sql`

func TestLogoutAllHandler(t *testing.T) {
	type testReq struct {
		uid any
	}

	type wantResp struct {
		code          int
		cookieDeleted bool
		isRevoked     bool
	}

	tests := []struct {
		name string
		req  testReq
		want wantResp
	}{
		{
			name: "#01_OK",
			req: testReq{
				uid: 1,
			},
			want: wantResp{
				code:          http.StatusNoContent,
				cookieDeleted: true,
				isRevoked:     true,
			},
		},
		{
			name: "#02_NO_UID",
			req:  testReq{},
			want: wantResp{
				code:          http.StatusInternalServerError,
				cookieDeleted: false,
			},
		},
		{
			name: "#03_WRONG_UID_TYPE",
			req: testReq{
				uid: "lmao",
			},
			want: wantResp{
				code:          http.StatusInternalServerError,
				cookieDeleted: false,
			},
		},
		{
			name: "#04_NO_TOKENS",
			req: testReq{
				uid: 41,
			},
			want: wantResp{
				code:          http.StatusNoContent,
				cookieDeleted: true,
			},
		},
	}

	handler := logoutAllHandler(svcAuth)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, testutil.LoadFixtures(pool, logoutAllFixtures, resetALLFixtures))
			req := httptest.NewRequest(
				http.MethodPost,
				"/auth/logoutAll",
				nil,
			)

			if tt.req.uid != nil {
				ctx := context.WithValue(
					req.Context(),
					middlewares.UIDKey,
					tt.req.uid,
				)
				req = req.WithContext(ctx)
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.want.code, rec.Code)
			assert.Equal(
				t,
				tt.want.cookieDeleted,
				refreshCookieDeleted(rec),
			)
			uid, ok := tt.req.uid.(int)
			if ok {
				assert.Equal(t, tt.want.isRevoked, checkUserRefreshRevoke(t, uid))
			}
		})
	}
}

func checkUserRefreshRevoke(t *testing.T, uid int) bool {
	t.Helper()

	const query = `
		SELECT
			EXISTS (
				SELECT 1
				FROM refresh_token
				WHERE user_id = $1
			)
			AND NOT EXISTS (
				SELECT 1
				FROM refresh_token
				WHERE user_id = $1
				  AND revoked IS NOT TRUE
			)`
	var revoked bool
	err := pool.QueryRow(
		context.Background(),
		query,
		uid,
	).Scan(&revoked)
	require.NoError(t, err)

	return revoked
}
