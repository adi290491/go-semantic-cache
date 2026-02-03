package database

import (
	"github.com/adi290491/semantic-cache/config"
	"github.com/redis/go-redis/v9"
)

type redisClient struct {
	redisCli *redis.Client
}

func NewRedisClient(cfg *config.Config) *redisClient {
	redisCfg := cfg.GetRedisConfig()

	return &redisClient{
		redisCli: redis.NewClient(&redis.Options{
			Addr:     redisCfg.Hostname + ":" + redisCfg.Port,
			Username: redisCfg.Username,
			Password: redisCfg.Password,
			DB:       0,
		}),
	}
}
