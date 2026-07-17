package httpserver

import (
	"fmt"
	"net/http"
	authsvc "notes/internal/gateway/service/auth"
)

type refreshDTO struct {
	Refresh string `json:"refresh"`
}

func refreshHandler(svctoken authsvc.IAuthSVC) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := getCtxLogger(ctx)

		var dtoRefresh refreshDTO
		if err := decodeJSON(&dtoRefresh, r); err != nil {
			handleError(w, fmt.Errorf("decode: %w", err), logger)
			return
		}

		tokens, err := svctoken.Refresh(ctx, dtoRefresh.Refresh)
		if err != nil {
			handleError(w, fmt.Errorf("svc.refresh: %w", err), logger)
			return
		}

		resp := AuthResp{
			Access:  tokens.Access,
			Refresh: tokens.Refresh,
		}
		writeJSON(w, http.StatusOK, logger, resp)
	}
}
