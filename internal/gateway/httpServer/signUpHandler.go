package httpserver

import (
	"fmt"
	"net/http"
	authsvc "notes/internal/gateway/service/auth"
)

type SignUpDTO struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

func signUpHandler(authSVC authsvc.IAuthSVC) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := getCtxLogger(ctx)

		var dtoUser SignUpDTO
		if err := decodeJSON(&dtoUser, r); err != nil {
			handleError(w, fmt.Errorf("decode: %w", err), logger)
			return
		}

		req := authsvc.SignUpReq{
			Login:    dtoUser.Login,
			Password: dtoUser.Password,
			Email:    dtoUser.Email,
		}

		tokens, err := authSVC.Signup(ctx, req)
		if err != nil {
			handleError(w, fmt.Errorf("SignUp: %w", err), logger)
			return
		}
		var cookie *http.Cookie
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
		writeJSON(w, http.StatusCreated, logger, resp)
	}
}
