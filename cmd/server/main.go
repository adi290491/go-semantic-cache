package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/adi290491/semantic-cache/config"
	"github.com/adi290491/semantic-cache/internal/ai"
	"github.com/adi290491/semantic-cache/internal/database"
	"github.com/adi290491/semantic-cache/internal/handler"
	"github.com/adi290491/semantic-cache/internal/router"
)

func init() {
	setupLogging()
}

func setupLogging() {

	var logger *slog.Logger
	logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	slog.SetDefault(logger)
}

func main() {

	cfg, err := config.LoadConfig()

	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	redisClient := database.NewRedisClient(cfg)
	aiClient := ai.NewQueryHandler(cfg)
	h := handler.NewHandler(redisClient, aiClient)

	mux := router.NewRouter(redisClient, aiClient, h)

	s := &http.Server{
		Addr:         ":" + cfg.Port,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      mux,
	}

	// go func() {
	// 	slog.Error("Server listening", "port", cfg.Port)
	// 	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
	// 		slog.Error("Server failed", "error", err)
	// 		os.Exit(1)
	// 	}
	// }()
	slog.Info("Server listening", "port", cfg.Port)
	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}

	// prompt := "Who is the CEO of Apple?"
	// handler.HandleUserQuery(prompt)

	// time.Sleep(5 * time.Second)
	// handler.HandleUserQuery(prompt)

	// time.Sleep(5 * time.Second)
	// prompt = "Do you know who the CEO of Apple is?"
	// handler.HandleUserQuery(prompt)

	// time.Sleep(5 * time.Second)
	// prompt = "Do you know who the CEO of Google is?"
	// handler.HandleUserQuery(prompt)

	// time.Sleep(5 * time.Second)
	// prompt = "What are the 8 planets of the solar system?"
	// handler.HandleUserQuery(prompt)

}
