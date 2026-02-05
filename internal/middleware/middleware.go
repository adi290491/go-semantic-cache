package middleware

import (
	"context"
	"encoding/json"
	"errors"
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

func NewCachingMiddleware(rdb *database.RedisClient, ai *ai.OpenAIHandler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
			defer cancel()
			// ctx := r.Context()

			var queryReq model.QueryRequestModel
			if err := json.NewDecoder(r.Body).Decode(&queryReq); err != nil {
				slog.Error("Failed to decode query request", "error", err)
				httperr.RespondWithError(w, fmt.Errorf("Invalid request body"), http.StatusBadRequest)
				return
			}

			queryHash := util.HashQuery(queryReq.Query)
			cacheKey := fmt.Sprintf("cache:%s", queryHash)

			// Find exact match
			slog.Info("Finding exact match")
			responseModel, err := rdb.Exists(ctx, cacheKey)
			if err == nil {
				slog.Info("Exact Cache hit")
				httperr.WriteJSON(w, responseModel, http.StatusOK)
				return
			}

			if !errors.Is(err, database.ErrCacheMiss) {
				slog.Error("Redis error while checking cache", "error", err)
				httperr.RespondWithError(w, fmt.Errorf("error while checking cache: %w", err), http.StatusInternalServerError)
				return
			}

			slog.Warn("Cache miss, fetching similar query")

			// Find similar match
			embedding, err := ai.GenerateEmbedding(ctx, queryReq.Query)

			if err != nil {
				slog.Error("Embedding error", "error", err)
				httperr.RespondWithError(w, fmt.Errorf("failed to generate embedding: %w", err), http.StatusInternalServerError)
				return
			}

			responseModel, err = rdb.FindSimilar(ctx, embedding)

			if err == nil && responseModel != nil {
				slog.Info("Similar Cache hit")
				httperr.WriteJSON(w, responseModel, http.StatusOK)
				return
			}

			if err != nil {
				slog.Warn("Error while fetching similar query", "error", err)
				// os.Exit(1)
			}

			slog.Info("Cache miss, calling handler")

			ctx = context.WithValue(ctx, "query", queryReq.Query)
			ctx = context.WithValue(ctx, "embedding", embedding)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
