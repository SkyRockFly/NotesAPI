package httpserver

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"notes/internal/pkg/apperror"
	"notes/internal/pkg/testutil"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	fixtureRefreshHandler = "testdata/fixtures/refresh_test/rows.sql"
	resetALLFixtures      = `TRUNCATE TABLE refresh_token,app_user RESTART IDENTITY CASCADE`
)

type authBodyMode int

const (
	bodyRaw authBodyMode = iota
	bodyAccess
	bodyAccessAndRefresh
)

func Test_refreshHandler(t *testing.T) {
	type wantReq struct {
		cookieBody      string
		cookieExpiresAt time.Time
		mode            authBodyMode
		hasCookie       bool
	}
	type wantResp struct {
		code int
		body string
	}
	tests := []struct {
		name  string
		req   wantReq
		want  wantResp
		check func(t *testing.T, sut http.Handler, body string)
	}{
		{
			name: "#01_OK",
			req: wantReq{
				cookieBody:      "18f11757-c9cf-47f4-a187-ddfda409abb4.JZO-pIBrPciCvkUdarnkUWzitwphl3rU5P0xFDPHEo4",
				cookieExpiresAt: time.Now().UTC().Add(time.Hour * 24 * 30),
				mode:            bodyAccess,
				hasCookie:       true,
			},
			want: wantResp{
				code: http.StatusOK,
				body: "OK",
			},
		},
		{
			name: "#02_BAD_REFRESH",
			req: wantReq{
				cookieBody: "a30369f9-7ba1-48c3-bb01efac3lGt]]]]izZ3WE+kUOyrWvsgIQRlEvbI=",
				mode:       bodyRaw,
				hasCookie:  true,
			},
			want: wantResp{
				code: http.StatusBadRequest,
				body: `{"error":"bad request"}`,
			},
		},
		{
			name: "#03_INVALID_FORMAT",
			req: wantReq{
				cookieBody: "a30369f9-7ba.1-48c3-bb01efac.3lGta9e8YBV8.Q/s7vVpizZ3WE+kU.OyrWvsgIQRlEvbI=",
				mode:       bodyRaw,
				hasCookie:  true,
			},
			want: wantResp{
				code: http.StatusBadRequest,
				body: `{"error":"bad request"}`,
			},
		},
		{
			name: "#04_INVALID_PRIVATE",
			req: wantReq{
				cookieBody: "18f11757-c9cf-47f4-a187-ddfda409abb4.JZO-pIBdarnkUWzitwphl3rU5P0xFDPHEo4",
				mode:       bodyRaw,
				hasCookie:  true,
			},
			want: wantResp{
				code: http.StatusUnauthorized,
				body: `{"error":"unauthorized"}`,
			},
		},
		{
			name: "#05_REVOKED",
			req: wantReq{
				cookieBody:      "8c0e3c78-c172-4b5c-b2fe-cc0cd2d795a4.lKTh7FUoTpNSkWm477scWKwhnf4CGRosWNX3YcjEH-o",
				cookieExpiresAt: time.Now().UTC().Add(time.Hour * 24 * 30),
				mode:            bodyRaw,
				hasCookie:       true,
			},
			want: wantResp{
				code: http.StatusUnauthorized,
				body: `{"error":"unauthorized"}`,
			},
		},
		{
			name: "#06_EXPIRED",
			req: wantReq{
				cookieBody:      "00f30850-3ceb-4daf-bd84-33d67e56b775.SdbVTVauiZLbe5JKDbiIMF-qeDvtD4b6F55rj0aMKrk",
				cookieExpiresAt: time.Now().UTC().Add(time.Hour * 24 * 2),
				mode:            bodyRaw,
				hasCookie:       true,
			},
			want: wantResp{
				code: http.StatusUnauthorized,
				body: `{"error":"unauthorized"}`,
			},
		},
		{
			name: "#07_NO_COOKIE",
			req: wantReq{
				mode:      bodyRaw,
				hasCookie: false,
			},
			want: wantResp{
				code: http.StatusUnauthorized,
				body: `{"error":"unauthorized"}`,
			},
		},
		{
			name: "#08_CHECK_REVOKE",
			req: wantReq{
				cookieBody:      "0049dd76-af66-490d-ac5a-d1d2ac8dfb0d.mZYBwwU-fpi-r9L1zPkXf69A-yz2ar1yy-6Pwilhksw",
				cookieExpiresAt: time.Now().UTC().Add(time.Hour * 24 * 19),
				mode:            bodyRaw,
				hasCookie:       true,
			},
			check: func(t *testing.T, sut http.Handler, body string) {
				method := http.MethodPost
				hndURL := "/auth/refresh"
				req := httptest.NewRequest(method, hndURL, nil)
				req.AddCookie(makeCookie("0049dd76-af66-490d-ac5a-d1d2ac8dfb0d.mZYBwwU-fpi-r9L1zPkXf69A-yz2ar1yy-6Pwilhksw",
					time.Now().UTC().Add(time.Hour*24*19)))
				rr := httptest.NewRecorder()
				sut.ServeHTTP(rr, req)
				assert.Equal(t, http.StatusUnauthorized, rr.Code)
				assert.Equal(t, `{"error":"unauthorized"}`, testutil.NormalizeJSON(t, rr.Body.String()))
			},
			want: wantResp{
				code: http.StatusUnauthorized,
				body: `{"error":"unauthorized"}`,
			},
		},
		{
			name: "#09_NOT_EXIST",
			req: wantReq{
				cookieBody:      "0049dd76-af66-490d-ac1a-d1d2ac8dfp0d.mZYBwwU-fpi-r9L1zPkXf69A-yz2ar1yy-6Pwilhksw",
				cookieExpiresAt: time.Now().UTC().Add(time.Hour * 24 * 2),
				mode:            bodyRaw,
				hasCookie:       true,
			},
			want: wantResp{
				code: http.StatusUnauthorized,
				body: `{"error":"unauthorized"}`,
			},
		},
	}

	sut := refreshHandler(svcAuth)
	method := http.MethodPost
	hndURL := "/auth/refresh"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, testutil.LoadFixtures(pool, fixtureRefreshHandler, resetALLFixtures))
			req := httptest.NewRequest(method, hndURL, nil)
			if tt.req.hasCookie {
				req.AddCookie(makeCookie(tt.req.cookieBody, tt.req.cookieExpiresAt))
			}

			rr := httptest.NewRecorder()
			sut.ServeHTTP(rr, req)

			body := parseResponse(t, rr, tt.req.mode)
			if tt.check == nil {
				assert.Equal(t, tt.want.code, rr.Code)
				assert.Equal(t, testutil.NormalizeJSON(t, tt.want.body),
					testutil.NormalizeJSON(t, body))
			} else {
				tt.check(t, sut, tt.req.cookieBody)
			}
		})
	}
}

func parseResponse(t *testing.T, rec *httptest.ResponseRecorder, mode authBodyMode) string {
	t.Helper()

	raw := rec.Body.Bytes()

	if mode == bodyRaw {
		return string(raw)
	}

	var resp AuthResp
	if err := json.Unmarshal(raw, &resp); err != nil {
		t.Logf("unmarshal auth response: %v", err)
		return string(raw)
	}

	if err := checkAccess(resp.Access); err != nil {
		t.Logf("check access: %v", err)
		return string(raw)
	}

	if mode == bodyAccessAndRefresh {
		var refresh string

		for _, cookie := range rec.Result().Cookies() {
			if cookie.Name == "refresh_token" {
				refresh = cookie.Value
				break
			}
		}

		if refresh == "" {
			t.Log("refresh cookie not found")
			return string(raw)
		}

		if err := parseRefresh(refresh); err != nil {
			t.Logf("check refresh: %v", err)
			return string(raw)
		}
	}

	return "OK"
}

func checkAccess(access string) error {
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{
			jwt.SigningMethodHS256.Alg(),
		}),
	)

	var claims jwt.RegisteredClaims

	if _, err := parser.ParseWithClaims(
		access,
		&claims,
		func(*jwt.Token) (any, error) {
			return []byte("le hard secret"), nil
		},
	); err != nil {
		return fmt.Errorf("parse access token: %w", err)
	}

	return nil
}

func parseRefresh(refresh string) error {
	parts := strings.Split(refresh, ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("%w: invalid refresh format", apperror.ErrBadRequest)
	}
	if _, err := base64.RawURLEncoding.DecodeString(parts[1]); err != nil {
		return fmt.Errorf("%w: invalid private part", apperror.ErrBadRequest)
	}

	return nil
}

func makeCookie(value string, expiresAt time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     "refresh_token",
		Value:    value,
		Path:     "/auth",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt,
	}
}
