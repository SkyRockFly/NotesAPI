package main

import (
	"NotesService/internal/db"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	httpserver "NotesService/internal/httpServer"
	notes "NotesService/internal/notesRepository"

	"github.com/joho/godotenv"
)

func main() {
	var err error
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := godotenv.Load("configs/app.env"); err != nil {
		stop()
		log.Fatalf("Failed to load env file:%v", err)
	}
	DBUrl := os.Getenv("DB_URL")

	pool, err := db.InitDB(ctx, DBUrl)
	if err != nil {
		stop()
		log.Fatalf("Failed to connect to DB:%v", err)
	}

	repo := notes.PostgresNewRepository(pool)

	service := notes.NewRepositoryImpl(repo)
	serverPort := os.Getenv("PORT")

	httpserver.StartServer(ctx, serverPort, service)
	fmt.Println("Server")
	<-ctx.Done()
	fmt.Println("bye,bye, медведи")
	stop()
}
