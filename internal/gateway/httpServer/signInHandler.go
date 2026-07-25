package httpserver

import (
	"fmt"
	"net/http"
	authsvc "notes/internal/gateway/service/auth"
)

type SignInDTO struct {
	Login      string `json:"login"`
	Password   string `json:"password"`
	RememberMe bool   `json:"remember_me"`
}

func signInHandler(auth authsvc.IAuthSVC) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := getCtxLogger(ctx)

		var dtoUser SignInDTO
		if err := decodeJSON(&dtoUser, r); err != nil {
			handleError(w, fmt.Errorf("decode: %w", err), logger)
			return
		}

		loginReq := authsvc.LoginReq{
			Login:      dtoUser.Login,
			Password:   dtoUser.Password,
			RememberMe: dtoUser.RememberMe,
		}

		tokens, err := auth.Login(ctx, loginReq)
		if err != nil {
			handleError(w, fmt.Errorf("svc.list: %w", err), logger)
			return
		}

		var cookie *http.Cookie
		fmt.Println(tokens)
		if tokens.Refresh != "" {
			cookie = &http.Cookie{
				Name:     "refresh_token",
				Value:    tokens.Refresh,
				Path:     "/auth",
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
				Expires:  tokens.RefreshExpiresAt,
			}
		}
		http.SetCookie(w, cookie)

		resp := AuthResp{
			Access: tokens.Access,
		}
		writeJSON(w, http.StatusOK, logger, resp)
	}
}
