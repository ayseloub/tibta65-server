package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"

	"github.com/Tibta65web/tibta65-server/internal/config"
	"github.com/Tibta65web/tibta65-server/internal/domain/achievement"
	"github.com/Tibta65web/tibta65-server/internal/domain/auth"
	"github.com/Tibta65web/tibta65-server/internal/domain/backgroundcontent"
	"github.com/Tibta65web/tibta65-server/pkg/database"
	"github.com/Tibta65web/tibta65-server/pkg/logger"

	_ "github.com/Tibta65web/tibta65-server/docs"
	echoSwagger "github.com/swaggo/echo-swagger"

	"github.com/Tibta65web/tibta65-server/pkg/storage"
)

const jwtExpiry = 2 * time.Hour

// @title Tibta65 API
// @version 1.0
// @description API backend untuk Tibta65 — auth admin, member, kegiatan, dll.
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Ketik "Bearer" diikuti spasi dan JWT token. Contoh: "Bearer eyJhbGc..."
func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	logger.Init(cfg.AppEnv == "development")

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect database")
	}
	defer db.Close()

	var fileStorage storage.Storage

	switch cfg.StorageDriver {
	case "s3":
		s3Storage, err := storage.NewS3Storage(
			cfg.S3Endpoint,
			cfg.S3AccessKey,
			cfg.S3SecretKey,
			cfg.S3BucketName,
			true, // useSSL
		)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to init S3 storage")
		}
		fileStorage = s3Storage
	default:
		fileStorage = storage.NewLocalStorage(cfg.UploadDir, cfg.UploadURL)
	}

	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo, cfg.JWTSecret, jwtExpiry)
	authHandler := auth.NewHandler(authService)

	bgRepo := backgroundcontent.NewRepository(db)
	bgService := backgroundcontent.NewService(bgRepo)
	bgHandler := backgroundcontent.NewHandler(bgService)

	achievementRepo := achievement.NewRepository(db)
	achievementService := achievement.NewService(achievementRepo, fileStorage)
	achievementHandler := achievement.NewHandler(achievementService)

	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{
			"https://tibta65.vercel.app", // ganti sesuai domain Vercel kamu yang beneran
		},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}))

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	e.Static(cfg.UploadURL, cfg.UploadDir)

	auth.RegisterRoutes(e, authHandler, cfg.JWTSecret)
	backgroundcontent.RegisterRoutes(e, bgHandler, cfg.JWTSecret)
	achievement.RegisterRoutes(e, achievementHandler, cfg.JWTSecret)

	go func() {
		if err := e.Start(":" + cfg.AppPort); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("failed to start server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("failed to shutdown server gracefully")
	}

	log.Info().Msg("server stopped")
}
