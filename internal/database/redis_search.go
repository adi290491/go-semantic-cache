package database

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/adi290491/semantic-cache/config"
	"github.com/adi290491/semantic-cache/internal/model"
	"github.com/adi290491/semantic-cache/internal/util"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	rdb *redis.Client
}

func NewRedisClient(cfg *config.Config) *RedisClient {
	redisCfg := cfg.GetRedisConfig()

	slog.Info("Connecting to Redis server at port", "PORT", redisCfg.Port)
	redisCli := redis.NewClient(&redis.Options{
		Addr:     redisCfg.Hostname + ":" + redisCfg.Port,
		Username: redisCfg.Username,
		Password: redisCfg.Password,
		DB:       0,
	})

	pong, err := redisCli.Ping(context.Background()).Result()
	slog.Info(pong)
	if err != nil || pong != "PONG" {
		slog.Error("Failed to connect to redis client")
		os.Exit(1)
	}

	redisClient := &RedisClient{
		rdb: redisCli,
	}

	if err := redisClient.CreateVectorIndex(context.Background()); err != nil {
		slog.Error("Failed to create vector index", "error", err)
		os.Exit(1)
	}

	slog.Info("Vector index created successfully")

	return redisClient
}

func (c *RedisClient) CreateVectorIndex(ctx context.Context) error {
	cmd := c.rdb.Do(
		ctx,
		"FT.CREATE", "cache_idx",
		"ON HASH",
		"PREFIX", "1", "cache:",
		"SCHEMA",
		"embedding", "VECTOR", "HNSW", "6",
		"TYPE", "FLOAT32",
		"DIM", "1536",
		"DISTANCE_METRIC", "COSINE",
	)

	return cmd.Err()
}

func (c *RedisClient) CacheEmbedding(ctx context.Context, key string, vector []float32, query string, response string) error {
	vectorBytes := util.Float32SliceToBytes(vector)

	err := c.rdb.HSet(ctx, key, model.ResponseModel{
		Response:   response,
		Embedding:  vectorBytes,
		Created_at: time.Now(),
	}).Err()

	return err
}

func (c *RedisClient) Exists(ctx context.Context, key string) (*model.ResponseModel, error) {
	result, err := c.rdb.HMGet(
		ctx, key,
		"response", "query",
	).Result()

	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("Field %s not found in key %s\n", "response", key)
		} else {
			return nil, fmt.Errorf("Could not get hash field: %v", err)
		}
	}

	response, ok := result[0].(string)
	if !ok {
		return nil, fmt.Errorf("Response field type assertion error")
	}

	query, ok := result[1].(string)
	if !ok {
		return nil, fmt.Errorf("Query field type assertion error")
	}

	return &model.ResponseModel{
		Response: response,
		Query:    query,
	}, nil
}

func (c *RedisClient) FindSimilar(ctx context.Context, queryVector []float32) (*model.ResponseModel, error) {
	vectorBytes := util.Float32SliceToBytes(queryVector)

	query := "(*)=>[KNN num_neighbours @embedding $vector AS score]"
	cmd := c.rdb.Do(
		ctx,
		"FT.SEARCH", "cache_idx",
		query,
		"PARAMS", "2", "vec", vectorBytes,
		"DIALECTS", "2",
		"RETURN", "2", "response", "score",
	)

	result, err := cmd.Result()
	if err != nil {
		return nil, err
	}

	response, ok := result.(string)
	if !ok {
		return nil, fmt.Errorf("Response field type assertion error")
	}

	fmt.Printf("Response: %v", response)
	// query, ok := result[1].(string)
	// if !ok {
	// 	return nil, fmt.Errorf("Query field type assertion error")
	// }

	return nil, nil
}
