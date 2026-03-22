package session

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/openai/openai-go"

	"chatbot/server/internal/scenario"
)

// Session holds per-conversation state for a single chat session.
type Session struct {
	ID          string
	ScenarioID  string
	Messages    []openai.ChatCompletionMessageParamUnion
	TurnCount   int
	CanaryToken string
	CreatedAt   time.Time
}

// NewSession creates a new Session for the given scenario.
// It generates a unique session ID and canary token via crypto/rand.
// The cfg parameter is accepted for future validation; only ScenarioID is stored.
func NewSession(scenarioID string, _ scenario.ScenarioConfig) (*Session, error) {
	idBytes := make([]byte, 8)
	if _, err := rand.Read(idBytes); err != nil {
		return nil, fmt.Errorf("generate session id: %w", err)
	}

	canaryBytes := make([]byte, 16)
	if _, err := rand.Read(canaryBytes); err != nil {
		return nil, fmt.Errorf("generate canary token: %w", err)
	}

	return &Session{
		ID:          hex.EncodeToString(idBytes),
		ScenarioID:  scenarioID,
		Messages:    []openai.ChatCompletionMessageParamUnion{},
		TurnCount:   0,
		CanaryToken: hex.EncodeToString(canaryBytes),
		CreatedAt:   time.Now(),
	}, nil
}

// TurnLimitReached reports whether the session has reached or exceeded maxTurns.
func (s *Session) TurnLimitReached(maxTurns int) bool {
	return s.TurnCount >= maxTurns
}

// AddTurn appends a user+assistant message pair to Messages and increments TurnCount.
func (s *Session) AddTurn(userMsg, assistantMsg string) {
	s.Messages = append(s.Messages,
		openai.UserMessage(userMsg),
		openai.AssistantMessage(assistantMsg),
	)
	s.TurnCount++
}

// ApplySlidingWindow drops the oldest user+assistant pairs from Messages until
// the estimated token count is within tokenBudget.
// Token estimate: sum of len(content) / 4 (chars-to-tokens approximation).
// The system prompt is not in Messages — only user/assistant turns are stored here.
// At least the last turn (2 messages) is always retained.
func (s *Session) ApplySlidingWindow(tokenBudget int) {
	for len(s.Messages) > 2 && estimateTokens(s.Messages) > tokenBudget {
		// Drop the oldest pair (user + assistant = 2 messages).
		s.Messages = s.Messages[2:]
	}
}

// estimateTokens returns a rough token estimate for a slice of messages:
// sum of all content lengths divided by 4.
func estimateTokens(msgs []openai.ChatCompletionMessageParamUnion) int {
	total := 0
	for _, msg := range msgs {
		total += len(extractContent(msg))
	}
	return total / 4
}

// extractContent pulls the string content out of a message param union.
// Handles the simple string case used by UserMessage and AssistantMessage.
func extractContent(msg openai.ChatCompletionMessageParamUnion) string {
	if u := msg.OfUser; u != nil && u.Content.OfString.Valid() {
		return u.Content.OfString.Value
	}
	if a := msg.OfAssistant; a != nil && a.Content.OfString.Valid() {
		return a.Content.OfString.Value
	}
	if sys := msg.OfSystem; sys != nil && sys.Content.OfString.Valid() {
		return sys.Content.OfString.Value
	}
	return ""
}
