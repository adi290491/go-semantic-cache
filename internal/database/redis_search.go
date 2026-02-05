package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/adi290491/semantic-cache/config"
	"github.com/adi290491/semantic-cache/internal/model"
	"github.com/adi290491/semantic-cache/internal/util"
	"github.com/redis/go-redis/v9"
)

var ErrCacheMiss = errors.New("cache miss")

const (
	// SimilarityThreshold defines the maximum cosine distance for cache hits.
	// Lower values = stricter matching (0.0 = identical, 1.0 = completely different)
	// 0.2 means "similar enough" - queries with distance < 0.2 are considered matches
	similarityThreshold = 0.2

	EmbeddingDimension = 1536
)

type RedisClient struct {
	rdb *redis.Client
}

func NewRedisClient(cfg *config.Config) *RedisClient {
	redisCfg := cfg.GetRedisConfig()

	slog.Info("Connecting to Redis server at port", "PORT", redisCfg.Port)
	redisCli := redis.NewClient(&redis.Options{
		Addr:     redisCfg.Hostname + ":" + redisCfg.Port,
		Password: redisCfg.Password,
		DB:       redisCfg.Db,
	})

	pong, err := redisCli.Ping(context.Background()).Result()

	if err != nil {
		slog.Error("Failed to connect to redis client", "error", err)
		os.Exit(1)
	}

	if pong != "PONG" {
		slog.Error("Unexpected Redis ping response", "response", pong)
		os.Exit(1)
	}

	slog.Info("Successfully connected to Redis", "PONG", pong)

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
	_ = c.rdb.Do(ctx, "FT.DROPINDEX", "cache_idx").Err()

	cmd := c.rdb.Do(
		ctx,
		"FT.CREATE", "cache_idx",
		"ON", "HASH",
		"PREFIX", "1", "cache:",
		"SCHEMA",
		"embedding", "VECTOR", "HNSW", "6",
		"TYPE", "FLOAT32",
		"DIM", EmbeddingDimension,
		"DISTANCE_METRIC", "COSINE",
	)

	return cmd.Err()
}

func (c *RedisClient) CacheEmbedding(ctx context.Context, key string, vector []float32, query string, response string) error {
	vectorBytes := util.Float32SliceToBytes(vector)

	err := c.rdb.HSet(ctx, key, model.ResponseModel{
		Response:   response,
		Embedding:  vectorBytes,
		Query:      query,
		Created_at: time.Now(),
	}).Err()

	if err := c.rdb.Expire(ctx, key, 24*time.Hour); err != nil {
		slog.Warn("Failed to set expiration", "error", err, "key", key)
	}

	return err
}

func (c *RedisClient) Exists(ctx context.Context, key string) (*model.ResponseModel, error) {
	// result, err := c.rdb.HMGet(
	// 	ctx, key,
	// 	"response", "query",
	// ).Result()

	type response struct {
		Response string `redis:"response"`
		Query    string `redis:"query"`
	}

	var resp response
	err := c.rdb.HGetAll(
		ctx,
		key,
	).Scan(&resp)

	if err != nil {
		return nil, fmt.Errorf("failed to get hash: %w", err)
	}

	slog.Debug("HGetAll result", "response", resp.Response, "query", resp.Query)

	// fmt.Printf("Response is %v", resp)
	if resp.Response == "" || resp.Query == "" {
		slog.Debug("Empty Response - Cache miss", "Response", resp)
		return nil, ErrCacheMiss
	}

	return &model.ResponseModel{
		Response: resp.Response,
		Query:    resp.Query,
	}, nil
}

func (c *RedisClient) FindSimilar(ctx context.Context, queryVector []float32) (*model.ResponseModel, error) {
	vectorBytes := util.Float32SliceToBytes(queryVector)

	query := "(*)=>[KNN 5 @embedding $vec AS score]"
	cmd := c.rdb.Do(
		ctx,
		"FT.SEARCH", "cache_idx",
		query,
		"PARAMS", "2", "vec", vectorBytes,
		"DIALECT", "2",
		"RETURN", "2", "response", "score",
		"LIMIT", "0", "1",
	)

	result, err := cmd.Result()
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	slog.Debug("Vector search result", "total_results", result)

	resultMap, ok := result.(map[interface{}]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result format: %T", result)
	}

	totalResults, ok := resultMap["total_results"].(int64)
	if !ok {
		if totalInt, ok := resultMap["total_results"].(int); ok {
			totalResults = int64(totalInt)
		} else {
			return nil, fmt.Errorf("invalid total_results type: %T", resultMap["total_results"])
		}
	}

	if totalResults == 0 {
		return nil, fmt.Errorf("no similar entries found")
	}

	results, ok := resultMap["results"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid results type: %T", resultMap["results"])
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("empty results array")
	}

	// First result should be a map with document data
	firstResult, ok := results[0].(map[interface{}]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid result document type: %T", results[0])
	}

	slog.Debug("Processing first result", "result", firstResult)

	extraAttr, ok := firstResult["extra_attributes"].(map[interface{}]interface{})
	if !ok {
		return nil, fmt.Errorf("extra_attrigutes not found or invalid type: %T", firstResult["extra_attributes"])
	}

	responseText, ok := extraAttr["response"].(string)
	if !ok {
		return nil, fmt.Errorf("response type not found or invalid type")
	}

	scoreStr, ok := extraAttr["score"].(string)
	if !ok {
		return nil, fmt.Errorf("score field not found or wrong type")
	}

	score, err := strconv.ParseFloat(scoreStr, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid score format: %w", err)
	}

	if score > similarityThreshold {
		return nil, fmt.Errorf("similarity score too high: %f > %f", score, similarityThreshold)
	}

	slog.Info("Similar entry found", "score", score)

	return &model.ResponseModel{
		Response: responseText,
	}, nil
}
