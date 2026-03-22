package llm

import (
	"context"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// OpenAIClient wraps the openai-go SDK.
type OpenAIClient struct {
	client openai.Client
	model  string
}

// NewOpenAIClient creates an OpenAIClient.
// apiKey is the OpenAI API key; baseURL overrides the endpoint (pass "" for default).
// model is the model name (e.g. "gpt-4o-mini").
func NewOpenAIClient(apiKey, baseURL, model string) *OpenAIClient {
	opts := []option.RequestOption{option.WithAPIKey(apiKey)}
	if baseURL != "" {
		opts = append(opts, option.WithBaseURL(baseURL))
	}
	return &OpenAIClient{
		client: openai.NewClient(opts...),
		model:  model,
	}
}

// StreamChat starts a streaming chat completion and returns a channel of Chunks.
// The channel is closed after the Done chunk or an error chunk.
func (c *OpenAIClient) StreamChat(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion) (<-chan Chunk, error) {
	ch := make(chan Chunk)

	go func() {
		defer close(ch)

		stream := c.client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
			Model:    openai.ChatModel(c.model),
			Messages: messages,
		})

		for stream.Next() {
			event := stream.Current()
			if len(event.Choices) > 0 {
				delta := event.Choices[0].Delta.Content
				if delta != "" {
					ch <- Chunk{Content: delta}
				}
			}
		}

		if err := stream.Err(); err != nil {
			ch <- Chunk{Err: err}
			return
		}

		ch <- Chunk{Done: true}
	}()

	return ch, nil
}
