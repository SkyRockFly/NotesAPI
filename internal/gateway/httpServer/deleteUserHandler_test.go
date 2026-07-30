package httpserver

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"notes/internal/pkg/middlewares"
	"notes/internal/pkg/testutil"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const deleteUserFixtures = `testdata/fixtures/delete/rows.sql`

func TestDeleteUserHandler(t *testing.T) {
	type testReq struct {
		uid any
	}

	type wantResp struct {
		code           int
		userDeleted    bool
		refreshRevoked bool
		cookieDeleted  bool
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
				code:           http.StatusNoContent,
				userDeleted:    true,
				refreshRevoked: true,
				cookieDeleted:  true,
			},
		},
		{
			name: "#02_WRONG_UID_TYPE",
			req: testReq{
				uid: "lmao",
			},
			want: wantResp{
				code:           http.StatusInternalServerError,
				userDeleted:    false,
				refreshRevoked: false,
				cookieDeleted:  false,
			},
		},

		{
			name: "#03_NO_UID",
			req:  testReq{},
			want: wantResp{
				code:           http.StatusInternalServerError,
				userDeleted:    false,
				refreshRevoked: false,
				cookieDeleted:  false,
			},
		},
		{
			name: "#04_ALREADY_DELETED",
			req: testReq{
				uid: 2,
			},
			want: wantResp{
				code:           http.StatusNoContent,
				userDeleted:    true,
				refreshRevoked: false,
				cookieDeleted:  true,
			},
		},
		{
			name: "#05_NOT_FOUND",
			req: testReq{
				uid: 3,
			},
			want: wantResp{
				code:           http.StatusNoContent,
				userDeleted:    false,
				refreshRevoked: false,
				cookieDeleted:  true,
			},
		},
	}

	handler := middlewares.LogMiddleware(deleteUserHandler(svcAuth))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, testutil.LoadFixtures(pool, deleteUserFixtures, resetALLFixtures))

			req := httptest.NewRequest(
				http.MethodDelete,
				"/user/delete",
				nil,
			)

			if tt.req.uid != nil {
				ctx := context.WithValue(req.Context(),
					middlewares.UIDKey, tt.req.uid,
				)
				req = req.WithContext(ctx)
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			uid, ok := tt.req.uid.(int)
			if ok {
				assert.Equal(t, tt.want.userDeleted,
					checkUserDeleted(t, uid),
				)

				assert.Equal(
					t,
					tt.want.refreshRevoked,
					checkUserRefreshRevoke(t, uid),
				)
			}

			assert.Equal(t, tt.want.cookieDeleted,
				refreshCookieDeleted(rec),
			)
		})
	}
}

func getUID(t *testing.T, uid any) int {
	t.Helper()

	id, ok := uid.(int)
	require.True(t, ok)

	return id
}

func checkUserDeleted(t *testing.T, uid int) bool {
	t.Helper()

	const query = `
		SELECT deleted_at IS NOT NULL
		FROM app_user
		WHERE id = $1
	`

	var deleted bool
	err := pool.QueryRow(
		context.Background(),
		query,
		uid,
	).Scan(&deleted)

	if errors.Is(err, pgx.ErrNoRows) {
		return false
	}

	require.NoError(t, err)
	return deleted
}
