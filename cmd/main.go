package main

import (
	appConfig "NotesService/config"
	"NotesService/internal/db"
	applogger "NotesService/internal/logger"
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
	defer stop() //zerolog до init'a (default logger)

	appConfig, err := appConfig.GetAppConfig()
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

	httpserver.StartServer(ctx, appConfig.Server.Port, service)
}
