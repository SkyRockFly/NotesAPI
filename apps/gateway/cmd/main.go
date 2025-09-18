package main

import (
	"context"
	"fmt"
	"notes/apps/gateway/internal/config"
	"notes/pkg/applogger"
	"os/signal"
	"syscall"

	httpserver "notes/apps/gateway/internal/httpServer"
	usernoterepo "notes/apps/gateway/internal/userNoteRepo"
	userservice "notes/apps/gateway/internal/userService"

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

	repo := usernoterepo.NewHttpRepo(appConfig.Server.Base)
	service := userservice.NewService(repo)

	log.Info().Msg("Setup successful")

	if err := httpserver.StartServer(ctx, appConfig.Server.Port, service); err != nil {
		log.Panic().Err(fmt.Errorf("server: %w", err)).Msg("start app")
	}
	log.Info().Msg("server stopped gracefully")
}
