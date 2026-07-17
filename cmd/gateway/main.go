package main

import (
	"context"
	"fmt"
	"notes/internal/gateway/config"
	httpserver "notes/internal/gateway/httpServer"
	notehttp "notes/internal/gateway/repository/note/http"
	refreshtokenpg "notes/internal/gateway/repository/refreshToken/pg"
	userpg "notes/internal/gateway/repository/user/pg"
	authsvc "notes/internal/gateway/service/auth"
	notesvc "notes/internal/gateway/service/note"
	"notes/internal/gateway/service/refreshsvc"
	usersvc "notes/internal/gateway/service/user"
	"notes/internal/pkg/applogger"
	"notes/internal/pkg/db"
	"notes/internal/pkg/middlewares"
	"os/signal"
	"syscall"

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

	pgToken := refreshtokenpg.NewRepository(pool)
	svcToken := refreshsvc.NewService(pgToken, appConfig.Auth.JWTSecret)

	pgUser := userpg.NewRepository(pool)
	svcUser := usersvc.NewService(pgUser)

	httpNote := notehttp.NewHTTPRepo(notehttp.SvcHTTPCfg{
		Host: appConfig.NotesSvc.Host,
		Port: appConfig.NotesSvc.Port,
	})
	svcNote := notesvc.NewService(httpNote)
	svcAuth := authsvc.New(svcToken, svcUser)

	rlCfg := middlewares.RateLimiterParameters{
		Requests:        appConfig.RateLimiter.Requests,
		Period:          appConfig.RateLimiter.Period,
		Burst:           appConfig.RateLimiter.Burst,
		VisitorTTL:      appConfig.RateLimiter.VisitorTTL,
		CleanupInterval: appConfig.RateLimiter.CleanupInterval,
	}

	log.Info().Msg("Setup successful")

	opts := httpserver.ServerOpts{
		SVCnotes:       svcNote,
		SVCauth:        svcAuth,
		Secret:         appConfig.Auth.JWTSecret,
		Port:           appConfig.Server.Port,
		RateLimiterCfg: rlCfg,
	}

	if err := httpserver.StartServer(ctx, opts); err != nil {
		log.Panic().Err(fmt.Errorf("server: %w", err)).Msg("start app")
	}

	log.Info().Msg("server stopped gracefully")
}
