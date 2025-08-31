package main

import (
	"NotesService/internal/applogger"
	"NotesService/internal/config"
	"NotesService/internal/db"
	noterepository "NotesService/internal/noteRepository"
	noteservice "NotesService/internal/noteService"
	"context"
	"fmt"
	"os/signal"
	"syscall"

	httpserver "NotesService/internal/httpServer"

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

	pool, err := db.InitDB(ctx, appConfig.DB.URL)
	if err != nil {
		log.Panic().Err(fmt.Errorf("initDB: %w", err)).Msg("start app")
	}
	defer pool.Close()

	repo := noterepository.NewPostgres(pool)
	service := noteservice.NewService(repo)

	if err := httpserver.StartServer(ctx, appConfig.Server.Port, service); err != nil {
		log.Panic().Err(fmt.Errorf("server: %w", err)).Msg("start app")
	}
	log.Info().Msg("server stopped gracefully")
}
