package httpserver

import (
	"fmt"
	"net/http"
	authsvc "notes/internal/gateway/service/auth"
)

type signUpDTO struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

func signUpHandler(authSVC authsvc.IAuthSVC) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := getCtxLogger(ctx)

		var dtoUser signUpDTO
		if err := decodeJSON(&dtoUser, r); err != nil {
			handleError(w, fmt.Errorf("decode: %w", err), logger)
			return
		}

		req := authsvc.SignUpReq{
			Login:    dtoUser.Login,
			Password: dtoUser.Password,
			Email:    dtoUser.Email,
		}

		resp, err := authSVC.Signup(ctx, req)
		if err != nil {
			handleError(w, fmt.Errorf("SignUp: %w", err), logger)
			return
		}

		writeJSON(w, http.StatusCreated, logger, resp)
	}
}
