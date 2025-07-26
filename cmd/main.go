package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"NotesService/internal/db"
	httpserver "NotesService/internal/httpServer"
	notes "NotesService/internal/notesRepository"

	"github.com/joho/godotenv"
)

func main() {
	var err error
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := godotenv.Load("configs/app.env"); err != nil {
		log.Fatalf("Failed to load env file:%v", err)
	}
	dbUrl := os.Getenv("DB_URL")

	pool, err := db.InitDB(ctx, dbUrl)
	if err != nil {
		log.Fatalf("Failed to connect to DB:%v", err)
	}

	repo := notes.PostgresNewRepository(pool)

	service := notes.NotesNewRepositoryImpl(repo)
	serverPort := os.Getenv("PORT")

	httpserver.StartServer(ctx, serverPort, service)
	fmt.Println("Server")
	<-ctx.Done()
	fmt.Println("bye,bye, медведи")
	stop()
}
