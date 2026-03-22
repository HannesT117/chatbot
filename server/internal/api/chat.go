package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/openai/openai-go"

	"chatbot/server/internal/filter"
	"chatbot/server/internal/llm"
	"chatbot/server/internal/prompt"
	"chatbot/server/internal/scenario"
	"chatbot/server/internal/session"
)

type chatRequest struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
}

type sseEvent struct {
	Type    string `json:"type"`
	Content string `json:"content,omitempty"`
}

func writeSSEEvent(w http.ResponseWriter, ev sseEvent) {
	b, _ := json.Marshal(ev)
	fmt.Fprintf(w, "data: %s\n\n", b)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

// ChatHandler returns an http.HandlerFunc that handles POST /api/chat.
// It streams LLM responses via Server-Sent Events with deterministic filtering.
func ChatHandler(
	store session.SessionStore,
	scenarios map[string]scenario.ScenarioConfig,
	llmClient llm.Client,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Decode request body.
		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if req.SessionID == "" || req.Message == "" {
			http.Error(w, "session_id and message are required", http.StatusBadRequest)
			return
		}

		// 2. Get session from store.
		sess, err := store.Get(req.SessionID)
		if err != nil {
			http.Error(w, "session not found", http.StatusNotFound)
			return
		}

		// 3. Get scenario config.
		scenarioCfg, ok := scenarios[sess.ScenarioID]
		if !ok {
			http.Error(w, "scenario not found for session", http.StatusInternalServerError)
			return
		}

		// 5. Set SSE response headers before any output.
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("X-Accel-Buffering", "no")

		// 4. Input filter — check AFTER setting SSE headers so we can send blocked event.
		if filter.ContainsBlocked(req.Message, scenarioCfg.BlocklistTerms) {
			writeSSEEvent(w, sseEvent{Type: "blocked"})
			return
		}

		// 6. Build system prompt.
		systemPrompt := prompt.Build(scenarioCfg, sess.CanaryToken)

		// 7. Build messages for LLM.
		messages := make([]openai.ChatCompletionMessageParamUnion, 0, 1+len(sess.Messages)+1)
		messages = append(messages, openai.SystemMessage(systemPrompt))
		messages = append(messages, sess.Messages...)
		messages = append(messages, openai.UserMessage(req.Message))

		// 8. Call LLM streaming.
		chunks, err := llmClient.StreamChat(r.Context(), messages)
		if err != nil {
			writeSSEEvent(w, sseEvent{Type: "blocked"})
			return
		}

		// 9. Stream tokens, accumulate, and filter.
		var accumulated strings.Builder

		for chunk := range chunks {
			// On error: fail closed.
			if chunk.Err != nil {
				writeSSEEvent(w, sseEvent{Type: "blocked"})
				return
			}

			// On done: break out of loop.
			if chunk.Done {
				break
			}

			// Accumulate content.
			accumulated.WriteString(chunk.Content)

			// Stream token to client.
			writeSSEEvent(w, sseEvent{Type: "token", Content: chunk.Content})

			// Check filters mid-stream.
			accStr := accumulated.String()
			if filter.ContainsCanary(accStr, sess.CanaryToken) || filter.ContainsBlocked(accStr, scenarioCfg.BlocklistTerms) {
				writeSSEEvent(w, sseEvent{Type: "blocked"})
				return
			}
		}

		// 10. Final check on complete output.
		fullResponse := accumulated.String()
		if filter.ContainsCanary(fullResponse, sess.CanaryToken) || filter.ContainsBlocked(fullResponse, scenarioCfg.BlocklistTerms) {
			writeSSEEvent(w, sseEvent{Type: "blocked"})
			return
		}

		// 12. Save turn and send done.
		sess.AddTurn(req.Message, fullResponse)
		sess.ApplySlidingWindow(scenarioCfg.TokenBudget)
		store.Save(sess) //nolint:errcheck

		writeSSEEvent(w, sseEvent{Type: "done"})
	}
}
