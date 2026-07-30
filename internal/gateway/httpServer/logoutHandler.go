package httpserver

import (
	"errors"
	"fmt"
	"net/http"
	authsvc "notes/internal/gateway/service/auth"
	"time"
)

func logoutHandler(svc authsvc.IAuthSVC) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := getCtxLogger(ctx)

		reqCookie, err := r.Cookie("refresh_token")
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				deleteRefreshCookie(w)
				w.WriteHeader(http.StatusNoContent)
				return
			}
			handleError(w, fmt.Errorf("read refresh cookie %w", err), logger)
			return
		}

		req := authsvc.LogoutReq{
			Refresh: reqCookie.Value,
		}

		if err := svc.Logout(ctx, req); err != nil {
			handleError(w, fmt.Errorf("svc.logout: %w", err), logger)
			return
		}

		deleteRefreshCookie(w)
		w.WriteHeader(http.StatusNoContent)
	}
}

func deleteRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/auth",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(1, 0),
	})
}
