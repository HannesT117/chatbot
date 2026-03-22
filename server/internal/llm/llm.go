package llm

import (
	"context"

	"github.com/openai/openai-go"
)

// Chunk is a single streaming event from the LLM.
type Chunk struct {
	Content string // token text; empty for the final chunk
	Done    bool   // true on the last chunk (Content will be empty)
	Err     error  // non-nil on error
}

// Client streams chat completions.
type Client interface {
	StreamChat(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion) (<-chan Chunk, error)
}
