package router

import (
	"net/http"

	"github.com/adi290491/semantic-cache/internal/ai"
	"github.com/adi290491/semantic-cache/internal/database"
	"github.com/adi290491/semantic-cache/internal/handler"
	"github.com/adi290491/semantic-cache/internal/middleware"
)

func NewRouter(redisClient *database.RedisClient, aiClient *ai.OpenAIHandler, h *handler.Handler) http.Handler {
	mux := http.NewServeMux()

	cachingMiddleware := middleware.NewCachingMiddleware(redisClient, aiClient)

	mux.Handle("POST /query", cachingMiddleware(http.HandlerFunc(h.HandleUserQuery)))

	return mux
}
