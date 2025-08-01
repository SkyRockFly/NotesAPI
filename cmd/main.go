package main

import (
	appConfig "NotesService/configs"
	"NotesService/internal/db"
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

	appConfig := appConfig.GetAppConfig()

	pool, err := db.InitDB(ctx, appConfig.DB.URL)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to DB:%v", err))
	}
	defer pool.Close()

	repo := noterepository.NewPostgres(pool)
	service := noteservice.NewService(repo)

	httpserver.StartServer(ctx, appConfig.Server.Port, service)
	fmt.Println("Server")
	<-ctx.Done()
	fmt.Println("bye,bye, медведи")
	stop()
}
