package httpserver

import (
	"fmt"
	"net/http"
	authsvc "notes/internal/gateway/service/auth"
	"notes/internal/pkg/apperror"
)

func refreshHandler(svctoken authsvc.IAuthSVC) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := getCtxLogger(ctx)

		reqCookie, err := r.Cookie("refresh_token")
		if err != nil {
			handleError(w, fmt.Errorf("find refresh cookie %w", apperror.ErrUnauthorized), logger)
			return
		}

		tokens, err := svctoken.Refresh(ctx, reqCookie.Value)
		if err != nil {
			handleError(w, fmt.Errorf("svc.refresh: %w", err), logger)
			return
		}

		resp := AuthResp{
			Access: tokens.Access,
		}
		var cookie *http.Cookie
		if resp.Refresh != "" {
			cookie = &http.Cookie{
				Name:     "refresh_token",
				Value:    resp.Refresh,
				Path:     "/auth",
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
				Expires:  tokens.RefreshExpiresAt,
			}
		}

		http.SetCookie(w, cookie)
		writeJSON(w, http.StatusOK, logger, resp)
	}
}
