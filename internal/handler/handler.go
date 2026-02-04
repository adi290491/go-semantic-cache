package handler

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/adi290491/semantic-cache/config"
	"github.com/adi290491/semantic-cache/internal/ai"
	"github.com/adi290491/semantic-cache/internal/database"
	"github.com/adi290491/semantic-cache/internal/model"
	"github.com/adi290491/semantic-cache/internal/util"
)

type Handler struct {
	redisCli       *database.RedisClient
	aiQueryHandler *ai.QueryHandler
}

func NewHandler(cfg *config.Config) *Handler {

	rdb := database.NewRedisClient(cfg)
	queryHandler := ai.NewQueryHandler(cfg)
	return &Handler{
		redisCli:       rdb,
		aiQueryHandler: queryHandler,
	}
}

func (h *Handler) HandleUserQuery(userQuery string) *model.ResponseModel {

	// prompt := userQuery

	queryHash := util.CreateQueryHash(userQuery)

	var responseModel *model.ResponseModel
	var err error
	if responseModel, err = h.redisCli.Exists(context.Background(), fmt.Sprintf("cache:%s", queryHash)); err == nil {
		return responseModel
	}

	slog.Info("Key error", "error", err)
	queryVector := h.aiQueryHandler.GetQueryEmbedding(context.Background(), userQuery)

	if responseModel, err = h.redisCli.FindSimilar(context.Background(), queryVector); err == nil {
		return responseModel
	}

	// h.aiQueryHandler.GetQueryEmbedding(context.Background(), userQuery)
	h.aiQueryHandler.HandleAIQuery(userQuery)
	return nil
}
