package grpcserver

import (
	"context"
	"errors"
	"fmt"
	"net"
	notesv1 "notes/internal/grpc/generated"
	noteservice "notes/internal/notes-svc/noteService"
	"notes/internal/pkg/apperror"
	"notes/internal/pkg/interceptors"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var isShuttingDown atomic.Bool

type Server struct {
	notesv1.UnimplementedNotesServiceServer
	service *noteservice.Service
}

func NewServer(service *noteservice.Service) *Server {
	return &Server{
		service: service,
	}
}

func StartServer(
	ctx context.Context,
	port string,
	service *noteservice.Service,
) error {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("grpc listener: %w", err)
	}

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.LogInterceptor(),
		),
	)

	notesv1.RegisterNotesServiceServer(server, NewServer(service))

	errs, eCtx := errgroup.WithContext(ctx)
	errs.Go(func() error {
		if err := server.Serve(listener); err != nil {
			return fmt.Errorf("serve grpc: %w", err)
		}
		return nil
	})

	<-eCtx.Done()
	isShuttingDown.Store(true)

	shutdownCtx, done := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer done()

	stopped := make(chan struct{})

	go func() {
		server.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		// all rpc done

	case <-shutdownCtx.Done():
		log.Warn().
			Err(shutdownCtx.Err()).
			Msg("grpc graceful shutdown")

		server.Stop() // forced shutdown
		<-stopped
	}

	if err := errs.Wait(); err != nil {
		return fmt.Errorf("server: %w", err)
	}
	return nil
}

func getCtxLogger(ctx context.Context) *zerolog.Logger {
	if v := ctx.Value(interceptors.LoggerCtxKey); v != nil {
		if lg, ok := v.(zerolog.Logger); ok {
			return &lg
		}
	}
	logger := zerolog.Nop()
	return &logger
}

func handleGRPCError(err error, log *zerolog.Logger) error {
	var (
		code  codes.Code
		msg   string
		level = zerolog.InfoLevel
	)

	switch {
	case errors.Is(err, apperror.ErrBadRequest):
		code = codes.InvalidArgument
		msg = "bad request"

	case errors.Is(err, apperror.ErrNotFound):
		code = codes.NotFound
		msg = "not found"

	case errors.Is(err, apperror.ErrAlreadyExists):
		code = codes.AlreadyExists
		msg = "already exists"

	case errors.Is(err, apperror.ErrUnauthorized):
		code = codes.Unauthenticated
		msg = "unauthorized"

	case errors.Is(err, apperror.ErrBackend):
		level = zerolog.ErrorLevel
		code = codes.Unavailable
		msg = "service unavailable"

	default:
		level = zerolog.ErrorLevel
		code = codes.Internal
		msg = "service error"
	}

	log.WithLevel(level).
		Err(fmt.Errorf("response: %w", err)).
		Str("grpc_code", code.String()).
		Msg("request failed")

	return fmt.Errorf("err: %w", status.Error(code, msg))
}
