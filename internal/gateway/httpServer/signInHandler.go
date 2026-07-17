package httpserver

import (
	"fmt"
	"net/http"
	authsvc "notes/internal/gateway/service/auth"
)

type signInDTO struct {
	Login      string `json:"login"`
	Password   string `json:"password"`
	RememberMe bool   `json:"remember_me"`
}

func signInHandler(auth authsvc.IAuthSVC) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := getCtxLogger(ctx)

		var dtoUser signInDTO
		if err := decodeJSON(&dtoUser, r); err != nil {
			handleError(w, fmt.Errorf("decode: %w", err), logger)
			return
		}

		loginReq := authsvc.LoginReq{
			Login:      dtoUser.Login,
			Password:   dtoUser.Password,
			RememberMe: dtoUser.RememberMe,
		}

		resp, err := auth.Login(ctx, loginReq)
		if err != nil {
			handleError(w, fmt.Errorf("svc.list: %w", err), logger)
			return
		}

		writeJSON(w, http.StatusOK, logger, resp)
	}
}
