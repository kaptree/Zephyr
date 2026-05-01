package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"labelpro-server/internal/config"
	"labelpro-server/internal/database"
	"labelpro-server/internal/logger"
	"labelpro-server/internal/models"
	"labelpro-server/internal/router"
	"labelpro-server/internal/utils"

	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load("config.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}
	config.SetActive(cfg, "config.json")

	if err := logger.Init(
		cfg.Log.Level,
		cfg.Log.Format,
		cfg.Log.OutputDir,
		cfg.Log.MaxSizeMB,
		cfg.Log.MaxBackups,
		cfg.Log.MaxAgeDays,
		cfg.Log.Compress,
		cfg.Log.EnableConsole,
	); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting LabelPro Server...",
		zap.String("version", "1.0.0"),
		zap.String("mode", cfg.Server.Mode),
	)

	if err := database.InitPostgres(cfg); err != nil {
		logger.Fatal("Failed to connect to PostgreSQL", zap.Error(err))
	}
	defer database.ClosePostgres()

	if err := database.DB.AutoMigrate(
		&models.User{},
		&models.Department{},
		&models.RolePermission{},
		&models.Note{},
		&models.Tag{},
		&models.Template{},
		&models.NoteAssignee{},
		&models.NoteAttachment{},
		&models.WorkGroup{},
		&models.WorkGroupMember{},
		&models.CollaborationRoom{},
		&models.Reminder{},
		&models.LedgerEntry{},
		&models.AIConfig{},
		&models.ConfigFileHistory{},
		&models.AdminLog{},
		&models.OperationLog{},
	); err != nil {
		logger.Fatal("Failed to auto migrate database", zap.Error(err))
	}
	logger.Info("Database migration completed")

	if err := database.InitRedis(cfg); err != nil {
		logger.Warn("Failed to connect to Redis, continuing without cache", zap.Error(err))
	}
	defer database.CloseRedis()

	if err := utils.InitJWT(&cfg.JWT); err != nil {
		logger.Warn("Failed to initialize JWT, generating temporary keys", zap.Error(err))
	}

	engine := router.Setup(cfg)

	srv := &http.Server{
		Addr:         cfg.ServerAddr(),
		Handler:      engine,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeoutSeconds) * time.Second,
	}

	go func() {
		logger.Info("Server listening", zap.String("addr", cfg.ServerAddr()))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}
