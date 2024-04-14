package app

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/realPointer/banners/config"
	v1 "github.com/realPointer/banners/internal/controller/http/v1"
	"github.com/realPointer/banners/internal/repository"
	"github.com/realPointer/banners/internal/service"
	"github.com/realPointer/banners/pkg/httpserver"
	"github.com/realPointer/banners/pkg/logger"
	"github.com/realPointer/banners/pkg/postgres"
	"github.com/realPointer/banners/pkg/redis"
)

func Run() {
	// Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Logger
	l, err := logger.New(cfg.Log.Level)
	if err != nil {
		log.Fatalf("Logger error: %s", err)
	}

	// Postgres
	pg, err := postgres.New(cfg.PG.URL, postgres.MaxPoolSize(cfg.PG.PoolMax))
	if err != nil {
		l.Fatal().Err(err).Msg("app - Run - postgres.New")
	}
	defer pg.Close()

	// Redis
	rdb, err := redis.New(cfg.Redis.URL)
	if err != nil {
		l.Fatal().Err(err).Msg("app - Run - redis.New")
	}

	// Repositories
	repositories := repository.NewRepositories(l, pg, rdb)

	// Services dependencies
	deps := service.ServicesDependencies{
		Repositories: repositories,
		SignKey:      cfg.JWT.SignKey,
		TokenTTL:     cfg.JWT.TokenTTL,
		Salt:         cfg.JWT.Salt,
	}
	services := service.NewServices(l, deps)

	// HTTP Server
	handler := v1.NewRouter(l, services)
	httpServer := httpserver.New(handler, httpserver.Port(cfg.HTTP.Port))

	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		l.Info().Msg("app - Run - signal: " + s.String())
	case err = <-httpServer.Notify():
		l.Err(err).Msg("app - Run - httpServer.Notify")
	}

	// Shutdown
	err = httpServer.Shutdown()
	if err != nil {
		l.Err(err).Msg("app - Run - httpServer.Shutdown")
	}
}
