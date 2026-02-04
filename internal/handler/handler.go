package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/adi290491/semantic-cache/config"
	"github.com/adi290491/semantic-cache/httperr"
	"github.com/adi290491/semantic-cache/internal/ai"
	"github.com/adi290491/semantic-cache/internal/database"
	"github.com/adi290491/semantic-cache/internal/model"
	"github.com/adi290491/semantic-cache/internal/util"
)

type Handler struct {
	redisCli       *database.RedisClient
	aiQueryHandler *ai.OpenAIHandler
}

func NewHandler(cfg *config.Config) *Handler {

	rdb := database.NewRedisClient(cfg)
	queryHandler := ai.NewQueryHandler(cfg)
	return &Handler{
		redisCli:       rdb,
		aiQueryHandler: queryHandler,
	}
}

func (h *Handler) HandleUserQuery(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var queryReq model.QueryRequestModel
	if err := json.NewDecoder(r.Body).Decode(&queryReq); err != nil {
		slog.Error("Failed to decode query request", "error", err)
		httperr.RespondWithError(w, fmt.Errorf("Invalid request body"), http.StatusBadRequest)
		return
	}

	queryHash := util.HashQuery(queryReq.Query)
	cacheKey := fmt.Sprintf("cache:%s", queryHash)
	fmt.Printf("CacheKey: %s\n", cacheKey)

	// Find exact match
	slog.Info("Finding exact match")
	responseModel, err := h.redisCli.Exists(ctx, cacheKey)
	if err == nil {
		slog.Info("Cache hit", "Key", cacheKey, "Response Model", responseModel)
		httperr.WriteJSON(w, responseModel, http.StatusOK)
		return
	}

	if !errors.Is(err, database.ErrCacheMiss) {
		slog.Error("Redis error while checking cache", "error", err)
		httperr.RespondWithError(w, fmt.Errorf("error while checking cache: %w", err), http.StatusInternalServerError)
		return
	}

	slog.Warn("Cache miss, fetching similar query", "key", cacheKey)

	// Find similar match
	queryVector, err := h.aiQueryHandler.GenerateEmbedding(ctx, queryReq.Query)

	if err != nil {
		slog.Error("Embedding error", "error", err)
		httperr.RespondWithError(w, fmt.Errorf("failed to generate embedding: %w", err), http.StatusInternalServerError)
		return
	}

	responseModel, err = h.redisCli.FindSimilar(ctx, queryVector)

	if err == nil && responseModel != nil {
		slog.Info("Similar Cache hit", "Key", cacheKey, "Response Model", responseModel)
		httperr.WriteJSON(w, responseModel, http.StatusOK)
		return
	}

	if err != nil {
		slog.Warn("Error while fetching similar query", "error", err)
		// os.Exit(1)
	}

	slog.Warn("Cache miss, fetching from API", "key", cacheKey)

	response, err := h.aiQueryHandler.HandleAIQuery(queryReq.Query)

	if err != nil {
		slog.Error("Error while calling OPENAI API", "error", err)
		os.Exit(1)
	}

	slog.Info("Caching response", "response", response)

	go func() {
		if err := h.redisCli.CacheEmbedding(
			ctx,
			cacheKey,
			queryVector,
			queryReq.Query,
			response); err != nil {
			slog.Warn("Failed to cache response", "error", err)
		}
		slog.Info("Response cached successfully")
	}()
}
