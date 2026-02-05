package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	OpenaiAPIKey string
	redisConfig  RedisConfig
	Port         string
}

type RedisConfig struct {
	Hostname string
	Password string
	Port     string
	Db       int
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		slog.Debug("Error loading .env file")
		// os.Exit(1)
	}

	openaiAPIKey := os.Getenv("OPENAI_API_KEY")

	if err := validateAPIKey(openaiAPIKey); err != nil {
		return nil, fmt.Errorf("OPENAI_API_KEY validation error: %v", err)
	}

	db, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if db < 0 || db > 15 {
		return nil, fmt.Errorf("invalid database number: %v", err)
	}

	redisCfg := &RedisConfig{
		Hostname: os.Getenv("REDIS_HOSTNAME"),
		Password: os.Getenv("REDIS_PASSWORD"),
		Port:     os.Getenv("REDIS_PORT"),
		Db:       db,
	}

	if redisCfg.Hostname == "" || redisCfg.Port == "" {
		return nil, fmt.Errorf("missing required database configuration. Please ensure REDIS_HOSTNAME, REDIS_PORT, and REDIS_PASSWORD are set")
	}

	if redisCfg.Password == "" {
		slog.Warn("Redis password not set - connecting without authentication")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8002"
		slog.Debug("PORT not specified, using default", "port", port)
	}

	config := &Config{
		OpenaiAPIKey: openaiAPIKey,
		redisConfig:  *redisCfg,
		Port:         port,
	}

	return config, nil
}

func (c *Config) GetRedisConfig() *RedisConfig {
	return &c.redisConfig
}

func validateAPIKey(api_key string) error {

	if api_key == "" {
		return fmt.Errorf("no API key was found - please head over to the troubleshooting notebook in this folder to identify & fix!")
	}

	if !strings.HasPrefix(api_key, "sk-proj-") {
		return fmt.Errorf("an API key was found, but it doesn't start sk-proj-; please check you're using the right key - see troubleshooting notebook")
	}

	if strings.TrimSpace(api_key) != api_key {
		return fmt.Errorf("an API key was found, but it looks like it might have space or tab characters at the start or end - please remove them - see troubleshooting notebook")
	}

	slog.Info("API key found and looks good so far!")
	return nil
}
