// @title           Observatory Auth Service
// @version         1.0
// @description     OIDC-based authentication service for the Observatory platform.
// @contact.name    Lucio Carvalho
// @host            localhost:8081
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/luciocarvalhojr/observatory-auth-svc/internal/config"
	"github.com/luciocarvalhojr/observatory-auth-svc/internal/handler"
	"github.com/luciocarvalhojr/observatory-auth-svc/internal/middleware"
	"github.com/luciocarvalhojr/observatory-auth-svc/internal/oidc"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// ── Structured logging ────────────────────────────────────────────
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.With().Str("service", "auth-svc").Logger()

	// ── Config ────────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	if cfg.Env == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// ── Redis ─────────────────────────────────────────────────────────
	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatal().Err(err).Msg("invalid REDIS_URL")
	}
	rdb := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal().Err(err).Msg("redis unreachable")
	}
	log.Info().Str("url", cfg.RedisURL).Msg("redis connected")

	// ── OIDC Provider ─────────────────────────────────────────────────
	provider, err := oidc.New(context.Background(), cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialise OIDC provider")
	}
	log.Info().Str("issuer", cfg.OIDCIssuer).Msg("oidc provider ready")

	// ── TTL ───────────────────────────────────────────────────────────
	ttl, err := time.ParseDuration(cfg.JWTAccessTTL)
	if err != nil {
		log.Fatal().Err(err).Str("value", cfg.JWTAccessTTL).Msg("invalid JWT_ACCESS_TTL")
	}

	// ── Router ────────────────────────────────────────────────────────
	r := gin.New()
	r.Use(middleware.Logger(), middleware.Recovery())

	// Health (no auth required)
	health := handler.NewHealth(rdb)
	health.Register(&r.RouterGroup)

	// Auth routes
	auth := handler.NewAuth(cfg, provider, rdb, ttl)
	auth.Register(r.Group("/auth"))

	// ── Server ────────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start in background
	go func() {
		log.Info().Str("port", cfg.Port).Msg("auth-svc starting")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	// ── Graceful shutdown ─────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down gracefully")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("shutdown error")
	}

	rdb.Close()
	log.Info().Msg("auth-svc stopped")
}
