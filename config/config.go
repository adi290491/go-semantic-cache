package config

import (
	"fmt"
	"log/slog"
	"os"
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
	Username string
	Password string
	Port     string
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		slog.Debug("Error loading .env file")
		// os.Exit(1)
	}

	openaiAPIKey := os.Getenv("OPENAI_API_KEY")

	if err := validateAPIKey(openaiAPIKey); err != nil {
		slog.Error("OPENAI_API_KEY validation error", "error", err)
		os.Exit(1)
	}

	redisCfg := &RedisConfig{
		Hostname: os.Getenv("DB_HOSTNAME"),
		Username: os.Getenv("DB_USERNAME"),
		Password: os.Getenv("DB_PASSWORD"),
		Port:     os.Getenv("DB_PORT"),
	}

	if redisCfg.Hostname == "" || redisCfg.Username == "" || redisCfg.Password == "" || redisCfg.Port == "" {
		return nil, fmt.Errorf("missing required database configuration. Please ensure DB_HOSTNAME, DB_PORT, DB_NAME, DB_USERNAME, and DB_PASSWORD are set")
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
