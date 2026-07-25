package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	authsvc "notes/internal/gateway/service/auth"
	notesvc "notes/internal/gateway/service/note"
	"notes/internal/pkg/apperror"
	"notes/internal/pkg/middlewares"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

const (
	encodeError = `{"error":"encode failed"}`
)

var isShuttingDown atomic.Bool

type errorAPIResponse struct {
	Err string `json:"error"`
}

type healthResponse struct {
	Health bool `json:"health"`
}

type ResponseNote struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AuthResp struct {
	Access string `json:"access,omitempty"`
}

type ServerOpts struct {
	SVCnotes       *notesvc.Service
	SVCauth        authsvc.IAuthSVC
	Secret         []byte
	Port           int
	RateLimiterCfg middlewares.RateLimiterParameters
}

type MW func(http.HandlerFunc) http.HandlerFunc

func Pipe(h http.HandlerFunc, mws ...MW) http.HandlerFunc {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

func StartServer(ctx context.Context, opts ServerOpts) error {
	mux := http.NewServeMux()

	authMW := middlewares.Auth(opts.Secret, 20*time.Second)

	rlAuth, err := middlewares.NewRateLimiter(ctx, getIPKey, opts.RateLimiterCfg)
	if err != nil {
		return fmt.Errorf("init rate limiter IP: %w", err)
	}

	mux.HandleFunc("POST /auth/signup", Pipe(
		signUpHandler(opts.SVCauth),
		middlewares.LogMiddleware,
		rlAuth.RateLimitMiddleware,
		middlewares.DemandJSONHeaders,
		middlewares.JSONReqSizeMiddleware,
	))

	mux.HandleFunc("POST /auth/signin", Pipe(
		signInHandler(opts.SVCauth),
		middlewares.LogMiddleware,
		rlAuth.RateLimitMiddleware,
		middlewares.DemandJSONHeaders,
		middlewares.JSONReqSizeMiddleware,
	))

	mux.HandleFunc("POST /auth/refresh", Pipe(
		refreshHandler(opts.SVCauth),
		middlewares.LogMiddleware,
		rlAuth.RateLimitMiddleware,
	))

	rlUID, err := middlewares.NewRateLimiter(ctx, getUIDKey, opts.RateLimiterCfg)
	if err != nil {
		return fmt.Errorf("init rate limiter UID: %w", err)
	}

	mux.HandleFunc("POST /note/get", Pipe(
		getNoteHandler(opts.SVCnotes),
		middlewares.LogMiddleware,
		authMW,
		rlUID.RateLimitMiddleware,
		middlewares.DemandJSONHeaders,
		middlewares.JSONReqSizeMiddleware,
	))

	mux.HandleFunc("DELETE /note/delete", Pipe(
		deleteNoteHandler(opts.SVCnotes),
		middlewares.LogMiddleware,
		authMW,
		rlUID.RateLimitMiddleware,
		middlewares.DemandJSONHeaders,
		middlewares.JSONReqSizeMiddleware,
	))

	mux.HandleFunc("POST /note/list", Pipe(
		listNoteHandler(opts.SVCnotes),
		middlewares.LogMiddleware,
		authMW,
		rlUID.RateLimitMiddleware,
		middlewares.DemandJSONHeaders,
		middlewares.JSONReqSizeMiddleware,
	))

	mux.HandleFunc("PUT /note/update", Pipe(
		updateNoteHandler(opts.SVCnotes),
		middlewares.LogMiddleware,
		authMW,
		rlUID.RateLimitMiddleware,
		middlewares.DemandJSONHeaders,
		middlewares.JSONReqSizeMiddleware,
	))

	mux.HandleFunc("POST /note/create", Pipe(
		createNoteHandler(opts.SVCnotes),
		middlewares.LogMiddleware,
		authMW,
		rlUID.RateLimitMiddleware,
		middlewares.DemandJSONHeaders,
		middlewares.JSONReqSizeMiddleware,
	))

	handler := middlewares.CORS(mux,
		"http://localhost:5682")

	server := &http.Server{
		Addr:    ":" + strconv.Itoa(opts.Port),
		Handler: handler,
	}

	errs, eCtx := errgroup.WithContext(ctx)
	errs.Go(func() error {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("listenAndServe: %w", err)
		}
		return nil
	})

	<-eCtx.Done()
	isShuttingDown.Store(true)

	shutdownCtx, done := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer done()

	if err := server.Shutdown(shutdownCtx); err != nil &&
		!errors.Is(err, http.ErrServerClosed) &&
		!errors.Is(err, context.Canceled) {
		if errors.Is(err, context.DeadlineExceeded) {
			_ = server.Close()
		}
		log.Warn().Err(err).Msg("graceful shutdown")
	}

	if err := errs.Wait(); err != nil {
		return fmt.Errorf("server: %w", err)
	}
	return nil
}

func HealthCheckHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := getCtxLogger(r.Context())
		if isShuttingDown.Load() {
			writeJSON(w, http.StatusServiceUnavailable, logger, errorAPIResponse{Err: "shutting down"})
			logger.
				Warn().
				Msg("httpServer.healthcheckHandler: shutting down")
			return
		}
		resp := &healthResponse{Health: true}
		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			logger.
				Warn().
				Err(err).
				Msg("httpServer.healthcheckHandler: json.Encode failed")
			return
		}
	}
}

func writeJSON(w http.ResponseWriter, statusCode int, logger *zerolog.Logger, resp any) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(resp); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(encodeError))
		logger.Error().Err(fmt.Errorf("encode: %w", err)).Msg("write JSON")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if statusCode == http.StatusNoContent {
		return
	}
	if _, err := w.Write(buf.Bytes()); err != nil {
		logger.Error().Err(fmt.Errorf("write: %w", err)).Msg("write JSON")
		return
	}
}

func decodeJSON(str any, r *http.Request) error {
	if str == nil {
		return fmt.Errorf("decode: empty struct")
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(str); err != nil {
		return fmt.Errorf("%w, decode: %w", apperror.ErrInvalidJSON, err)
	}

	return nil
}

func remapSvcToRespNote(note notesvc.Note) ResponseNote {
	newNote := ResponseNote{
		ID:        note.ID,
		Title:     note.Title,
		Body:      note.Body,
		CreatedAt: note.CreatedAt,
		UpdatedAt: note.UpdatedAt,
	}

	return newNote
}

func getCtxLogger(ctx context.Context) *zerolog.Logger {
	if v := ctx.Value(middlewares.LoggerCtxKey); v != nil {
		if lg, ok := v.(zerolog.Logger); ok {
			return &lg
		}
	}
	logger := zerolog.Nop()
	return &logger
}

func handleError(w http.ResponseWriter, err error, log *zerolog.Logger) {
	var (
		code  int
		resp  errorAPIResponse
		level = zerolog.InfoLevel
	)
	switch {
	case errors.Is(err, apperror.ErrInvalidJSON):
		code = http.StatusUnprocessableEntity
		resp.Err = "invalid json"

	case errors.Is(err, apperror.ErrBadRequest):
		code = http.StatusBadRequest
		resp.Err = "bad request"

	case errors.Is(err, apperror.ErrNotFound):
		code = http.StatusNotFound
		resp.Err = "not found"

	case errors.Is(err, apperror.ErrBackend):
		level = zerolog.ErrorLevel
		code = http.StatusBadGateway
		resp.Err = "gateway error"

	case errors.Is(err, apperror.ErrAlreadyExists):
		code = http.StatusConflict
		resp.Err = "already exists"

	case errors.Is(err, apperror.ErrUnauthorized):
		code = http.StatusUnauthorized
		resp.Err = "unauthorized"
	default:
		level = zerolog.ErrorLevel
		code = http.StatusInternalServerError
		resp.Err = "service error"
	}

	log.WithLevel(level).
		Err(err).
		Msg("request failed")

	writeJSON(w, code, log, resp)
}

func getUIDKey(r *http.Request) (string, error) {
	ctx := r.Context()
	uid, ok := ctx.Value(middlewares.UIDKey).(int)
	if !ok {
		return "", fmt.Errorf("can't extract uid")
	}
	key := strconv.Itoa(uid)
	return key, nil
}

func getIPKey(r *http.Request) (string, error) {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host, nil
	}

	if ip := net.ParseIP(r.RemoteAddr); ip != nil {
		return ip.String(), nil
	}

	return "", fmt.Errorf("invalid RemoteAddr %q: %w", r.RemoteAddr, err)
}
