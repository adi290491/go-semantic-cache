package ai

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/adi290491/semantic-cache/config"
	openai "github.com/sashabaranov/go-openai"
)

type QueryHandler struct {
	client *openai.Client
}

func NewQueryHandler(cfg *config.Config) *QueryHandler {
	client := openai.NewClient(cfg.OpenaiAPIKey)
	return &QueryHandler{
		client: client,
	}
}

func (q *QueryHandler) HandleAIQuery(prompt string) {

	slog.Info("Calling OPENAI API")
	ctx := context.Background()

	// client := openai.NewClient(option.WithAPIKey(cfg.OpenaiAPIKey))

	// prompt := "Who is the CEO of Apple?"

	// resp, err := q.client.Responses.New(ctx, responses.ResponseNewParams{
	// 	Input: responses.ResponseNewParamsInputUnion{OfString: openai.String(prompt)},
	// 	Model: openai.ChatModelGPT4_1Mini,
	// })
	resp, err := q.client.CreateChatCompletion(
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
		panic(err)
	}

	fmt.Println(resp.Choices[0].Message.Content)
}

func (q *QueryHandler) GetQueryEmbedding(ctx context.Context, query string) []float32 {

	queryRes, err := q.client.CreateEmbeddings(ctx,
		openai.EmbeddingRequest{
			Model: openai.SmallEmbedding3,
			Input: []string{"How many chucks would a woodchuck chuck"},
		},
	)

	if err != nil {
		slog.Error("Error creating query embedding", "error", err)
		os.Exit(1)
	}

	// targetRes, err := q.client.CreateEmbeddings(
	// 	ctx,
	// 	openai.EmbeddingRequest{
	// 		Model: openai.SmallEmbedding3,
	// 		Input: []string{"How many chucks would a woodchuck chuck if the woodchuck could chuck wood"},
	// 	},
	// )

	// if err != nil {
	// 	slog.Error("Error creating target embedding:", err)
	// 	os.Exit(1)
	// }

	// fmt.Println("Query Response: %v", queryRes.Data[0])
	// queryEmbedding := queryRes.Data
	// targetEmbedding := targetRes.Data[0]

	// similarity, err := queryEmbedding.DotProduct(&targetEmbedding)
	// if err != nil {
	// 	slog.Error("Error calculating dot product:", "error", err)
	// }

	// fmt.Printf("Similarity between query and target is %f", similarity)
	// vectors := make([][]float32, 0)
	// for _, data := range queryRes.Data {
	// 	vectors = append(vectors, data.Embedding)
	// }
	// return vectors
	return queryRes.Data[0].Embedding
}
