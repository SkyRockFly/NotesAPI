package main

import (
	"context"
	"fmt"
	"notes/internal/gateway/config"
	"notes/internal/pkg/applogger"
	"os/signal"
	"syscall"

	httpserver "notes/internal/gateway/httpServer"
	usernoterepo "notes/internal/gateway/userNoteRepo"
	userservice "notes/internal/gateway/userService"

	"github.com/rs/zerolog/log"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	appConfig, err := config.GetAppConfig()
	if err != nil {
		log.Panic().Err(fmt.Errorf("load config: %w", err)).Msg("start app")
	}

	applogger.Configure(applogger.LoggerCfg{
		FormatTimestamp: appConfig.Log.Timestamp,
		FormatLevel:     appConfig.Log.FormatLevel,
		Loglvl:          appConfig.Log.Level,
	})

	repo := usernoterepo.NewHTTPRepo(usernoterepo.SvcHTTPCfg{
		Host: appConfig.NotesSvc.Host,
		Port: appConfig.NotesSvc.Port,
	})
	service := userservice.NewService(repo)

	log.Info().Msg("Setup successful")

	if err := httpserver.StartServer(ctx, appConfig.Server.Port, service); err != nil {
		log.Panic().Err(fmt.Errorf("server: %w", err)).Msg("start app")
	}
	log.Info().Msg("server stopped gracefully")
}
