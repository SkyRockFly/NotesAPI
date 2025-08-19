package main

import (
	appConfig "NotesService/configs"
	"NotesService/internal/db"
	applogger "NotesService/internal/logger"
	noterepository "NotesService/internal/noteRepository"
	noteservice "NotesService/internal/noteService"
	"context"
	"fmt"
	"os/signal"
	"syscall"

	httpserver "NotesService/internal/httpServer"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	appConfig, err := appConfig.GetAppConfig()
	if err != nil {
		panic(fmt.Errorf("failed to connect to DB: %w", err))
	}

	applogger.Configure()

	pool, err := db.InitDB(ctx, appConfig.DB.URL)
	if err != nil {
		panic(fmt.Errorf("initDB: %w", err))
	}
	defer pool.Close()

	repo := noterepository.NewPostgres(pool)
	service := noteservice.NewService(repo)

	httpserver.StartServer(ctx, appConfig.Server.Port, service)
	<-ctx.Done()
	stop()
}
