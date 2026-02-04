package main

import (
	"log/slog"
	"os"

	"github.com/adi290491/semantic-cache/config"
	"github.com/adi290491/semantic-cache/internal/handler"
)

func main() {
	// if err := godotenv.Load(); err != nil {
	// 	slog.Debug("Error loading .env file")
	// 	// os.Exit(1)
	// }
	// openaiAPIKey := os.Getenv("OPENAI_API_KEY")

	// if openaiAPIKey == "" {
	// 	slog.Error("OPENAI_API_KEY is not set")
	// 	os.Exit(1)
	// }

	// redis.NewClient(&redis.Options{
	// 	Addr: "localhost:6379",
	// })

	cfg, err := config.LoadConfig()

	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	handler := handler.NewHandler(cfg)

	prompt := "Who is the CEO of Apple?"
	handler.HandleUserQuery(prompt)

	prompt = "Remind me who the CEO of Apple is?"
	handler.HandleUserQuery(prompt)

	// rdb := database.NewRedisClient(cfg)
	// if rdb != nil {
	// 	slog.Info("Redis client created at port " + cfg.Port)
	// }

	// ctx := context.Background()

	// client := openai.NewClient(option.WithAPIKey(cfg.OpenaiAPIKey))

	// prompt := "Who is the CEO of Apple?"

	// resp, err := client.Responses.New(ctx, responses.ResponseNewParams{
	// 	Input: responses.ResponseNewParamsInputUnion{OfString: openai.String(prompt)},
	// 	Model: openai.ChatModelGPT4_1Mini,
	// })

	// if err != nil {
	// 	panic(err)
	// }

	// print(resp.OutputText())

	// queryHandler := ai.NewQueryHandler(cfg)

	// prompt := "Who is the CEO of Apple?"
	// queryHandler.HandleAIQuery(prompt)

}
