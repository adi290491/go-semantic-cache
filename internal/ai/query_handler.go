package ai

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/adi290491/semantic-cache/config"
	openai "github.com/sashabaranov/go-openai"
)

type OpenAIHandler struct {
	client *openai.Client
}

func NewQueryHandler(cfg *config.Config) *OpenAIHandler {
	client := openai.NewClient(cfg.OpenaiAPIKey)
	return &OpenAIHandler{
		client: client,
	}
}

func (h *OpenAIHandler) HandleAIQuery(prompt string) (string, error) {

	slog.Info("Calling OPENAI API")
	ctx := context.Background()

	// client := openai.NewClient(option.WithAPIKey(cfg.OpenaiAPIKey))

	// prompt := "Who is the CEO of Apple?"

	// resp, err := q.client.Responses.New(ctx, responses.ResponseNewParams{
	// 	Input: responses.ResponseNewParamsInputUnion{OfString: openai.String(prompt)},
	// 	Model: openai.ChatModelGPT4_1Mini,
	// })
	resp, err := h.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4Dot1Mini,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		return "", err
	}

	fmt.Println(resp.Choices[0].Message.Content)
	return resp.Choices[0].Message.Content, nil
}

func (h *OpenAIHandler) GenerateEmbedding(ctx context.Context, query string) ([]float32, error) {

	queryRes, err := h.client.CreateEmbeddings(ctx,
		openai.EmbeddingRequest{
			Model: openai.SmallEmbedding3,
			Input: []string{query},
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create query embedding: %w", err)
	}

	embedding := queryRes.Data[0].Embedding
	if len(embedding) == 0 {
		return nil, fmt.Errorf("empty embedding vector returned")
	}
	slog.Debug("Generated embedding", "dimenstion", len(embedding))

	return embedding, nil
}
