package httpserver

import (
	"fmt"
	"net/http"
	authsvc "notes/internal/gateway/service/auth"
	"notes/internal/pkg/middlewares"
)

func logoutAllHandler(svc authsvc.IAuthSVC) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := getCtxLogger(ctx)

		uid, ok := ctx.Value(middlewares.UIDKey).(int)
		if !ok {
			handleError(w, fmt.Errorf("cant extract ctx value"), logger)
			return
		}

		svcReq := authsvc.LogoutAllReq{
			UserID: uid,
		}
		if err := svc.LogoutAll(ctx, svcReq); err != nil {
			handleError(w, fmt.Errorf("svc.logoutAll: %w", err), logger)
			return
		}

		deleteRefreshCookie(w)
		w.WriteHeader(http.StatusNoContent)
	}
}
