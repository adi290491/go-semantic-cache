package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

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

func NewHandler(rdb *database.RedisClient, ai *ai.OpenAIHandler) *Handler {

	return &Handler{
		redisCli:       rdb,
		aiQueryHandler: ai,
	}
}

func (h *Handler) HandleUserQuery(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	queryReq, ok := ctx.Value(model.QueryKey).(string)
	if !ok || queryReq == "" {
		slog.Error("Query not found in context")
		httperr.RespondWithError(w, fmt.Errorf("invalid request state"), http.StatusInternalServerError)
		return
	}

	err := validateQuery(queryReq)
	if err != nil {
		httperr.RespondWithError(w, err, http.StatusBadRequest)
	}

	queryHash := util.HashQuery(queryReq)
	cacheKey := fmt.Sprintf("cache:%s", queryHash)
	slog.Debug("Generated cache key", "key", cacheKey)

	embedding, ok := ctx.Value(model.EmbeddingKey).([]float32)

	if !ok || embedding == nil {

		embedding, err = h.aiQueryHandler.GenerateEmbedding(ctx, queryReq)
		if err != nil {
			slog.Warn("Cannot cache without embedding")
		}
	}

	response, err := h.aiQueryHandler.HandleAIQuery(queryReq)

	if err != nil {
		slog.Error("Error while calling OPENAI API", "error", err)
		httperr.RespondWithError(w, fmt.Errorf("failed to generate response: %w", err), http.StatusInternalServerError)
		return
	}

	slog.Info("Caching response", "response", response)

	go func() {
		cacheCtx, cacheCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cacheCancel()

		if err := h.redisCli.CacheEmbedding(
			cacheCtx,
			cacheKey,
			embedding,
			queryReq,
			response); err != nil {
			slog.Warn("Failed to cache response", "error", err)
		}
		slog.Info("Response cached successfully")
	}()

	httperr.WriteJSON(w, model.ResponseModel{
		Response: response,
		Query:    queryReq,
	}, http.StatusOK)
}

func validateQuery(queryReq string) error {
	// Validate query
	if queryReq == "" {

		return fmt.Errorf("query cannot be empty")
	}

	if len(queryReq) > 4000 { // OpenAI token limits
		
		return fmt.Errorf("query too long (max 4000 characters)")
	}
	return nil
}
