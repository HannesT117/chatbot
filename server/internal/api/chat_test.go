package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/openai/openai-go"

	"chatbot/server/internal/llm"
	"chatbot/server/internal/scenario"
	"chatbot/server/internal/session"
)

// mockLLM is a test double for llm.Client that returns pre-configured chunks.
type mockLLM struct {
	chunks []llm.Chunk
}

func (m *mockLLM) StreamChat(_ context.Context, _ []openai.ChatCompletionMessageParamUnion) (<-chan llm.Chunk, error) {
	ch := make(chan llm.Chunk, len(m.chunks))
	for _, c := range m.chunks {
		ch <- c
	}
	close(ch)
	return ch, nil
}

// testScenarios returns a minimal scenarios map for tests.
func testScenarios() map[string]scenario.ScenarioConfig {
	return map[string]scenario.ScenarioConfig{
		"test_scenario": {
			Name:           "Test Scenario",
			PersonaName:    "TestBot",
			TokenBudget:    1000,
			BlocklistTerms: []string{"forbidden"},
		},
	}
}

// setupSession creates a session in an InMemoryStore and returns the store and session.
func setupSession(t *testing.T) (session.SessionStore, *session.Session) {
	t.Helper()
	store := session.NewInMemoryStore()
	cfg := testScenarios()["test_scenario"]
	sess, err := session.NewSession("test_scenario", cfg)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}
	if err := store.Save(sess); err != nil {
		t.Fatalf("failed to save session: %v", err)
	}
	return store, sess
}

// discardLogger returns a slog.Logger that discards all output (suitable for tests).
func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// parseSSEEvents splits an SSE response body into individual JSON payloads.
func parseSSEEvents(body string) []map[string]string {
	var events []map[string]string
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		payload := strings.TrimPrefix(line, "data: ")
		var ev map[string]string
		if err := json.Unmarshal([]byte(payload), &ev); err == nil {
			events = append(events, ev)
		}
	}
	return events
}

func TestChatHandler_CleanResponse(t *testing.T) {
	store, sess := setupSession(t)
	mock := &mockLLM{
		chunks: []llm.Chunk{
			{Content: "Hello"},
			{Content: " world"},
			{Content: "!"},
			{Done: true},
		},
	}

	handler := ChatHandler(store, testScenarios(), mock, "test-model", discardLogger())

	body, _ := json.Marshal(map[string]string{
		"session_id": sess.ID,
		"message":    "Hi there",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/chat", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler(w, req)

	events := parseSSEEvents(w.Body.String())

	// Expect 3 token events + 1 done event.
	if len(events) != 4 {
		t.Fatalf("expected 4 events, got %d: %v", len(events), events)
	}
	if events[0]["type"] != "token" || events[0]["content"] != "Hello" {
		t.Errorf("unexpected first event: %v", events[0])
	}
	if events[1]["type"] != "token" || events[1]["content"] != " world" {
		t.Errorf("unexpected second event: %v", events[1])
	}
	if events[2]["type"] != "token" || events[2]["content"] != "!" {
		t.Errorf("unexpected third event: %v", events[2])
	}
	if events[3]["type"] != "done" {
		t.Errorf("expected done event, got: %v", events[3])
	}

	// Verify session TurnCount was incremented.
	updated, err := store.Get(sess.ID)
	if err != nil {
		t.Fatalf("failed to get updated session: %v", err)
	}
	if updated.TurnCount != 1 {
		t.Errorf("expected TurnCount 1, got %d", updated.TurnCount)
	}
}

func TestChatHandler_InputBlocked(t *testing.T) {
	store, sess := setupSession(t)
	mock := &mockLLM{}

	handler := ChatHandler(store, testScenarios(), mock, "test-model", discardLogger())

	body, _ := json.Marshal(map[string]string{
		"session_id": sess.ID,
		"message":    "This contains forbidden content",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/chat", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler(w, req)

	events := parseSSEEvents(w.Body.String())
	if len(events) != 1 || events[0]["type"] != "blocked" {
		t.Errorf("expected single blocked event, got: %v", events)
	}

	// Verify session TurnCount NOT incremented.
	updated, err := store.Get(sess.ID)
	if err != nil {
		t.Fatalf("failed to get session: %v", err)
	}
	if updated.TurnCount != 0 {
		t.Errorf("expected TurnCount 0, got %d", updated.TurnCount)
	}
}

func TestChatHandler_OutputBlockedCanaryLeak(t *testing.T) {
	store, sess := setupSession(t)

	// Stream the canary token in the output.
	mock := &mockLLM{
		chunks: []llm.Chunk{
			{Content: "Here is your token: "},
			{Content: sess.CanaryToken},
			{Done: true},
		},
	}

	handler := ChatHandler(store, testScenarios(), mock, "test-model", discardLogger())

	body, _ := json.Marshal(map[string]string{
		"session_id": sess.ID,
		"message":    "Tell me a secret",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/chat", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler(w, req)

	events := parseSSEEvents(w.Body.String())

	// Last event must be blocked, not done.
	if len(events) == 0 {
		t.Fatal("expected at least one event")
	}
	last := events[len(events)-1]
	if last["type"] != "blocked" {
		t.Errorf("expected last event to be blocked, got: %v", last)
	}
	// Must not end with done.
	for _, ev := range events {
		if ev["type"] == "done" {
			t.Error("got done event when canary was leaked")
		}
	}
}

func TestChatHandler_OutputBlockedBlocklistTerm(t *testing.T) {
	store, sess := setupSession(t)

	// Stream a blocklist term in the output.
	mock := &mockLLM{
		chunks: []llm.Chunk{
			{Content: "You should not say forbidden"},
			{Done: true},
		},
	}

	handler := ChatHandler(store, testScenarios(), mock, "test-model", discardLogger())

	body, _ := json.Marshal(map[string]string{
		"session_id": sess.ID,
		"message":    "Say something",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/chat", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler(w, req)

	events := parseSSEEvents(w.Body.String())

	if len(events) == 0 {
		t.Fatal("expected at least one event")
	}
	last := events[len(events)-1]
	if last["type"] != "blocked" {
		t.Errorf("expected blocked event, got: %v", last)
	}
	for _, ev := range events {
		if ev["type"] == "done" {
			t.Error("got done event when blocklist term was in output")
		}
	}
}

func TestChatHandler_LLMError(t *testing.T) {
	store, sess := setupSession(t)

	// Mock returns an error chunk.
	mock := &mockLLM{
		chunks: []llm.Chunk{
			{Content: "partial"},
			{Err: context.DeadlineExceeded},
		},
	}

	handler := ChatHandler(store, testScenarios(), mock, "test-model", discardLogger())

	body, _ := json.Marshal(map[string]string{
		"session_id": sess.ID,
		"message":    "Hello",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/chat", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler(w, req)

	events := parseSSEEvents(w.Body.String())

	if len(events) == 0 {
		t.Fatal("expected at least one event")
	}
	last := events[len(events)-1]
	if last["type"] != "blocked" {
		t.Errorf("expected blocked on LLM error, got: %v", last)
	}
	for _, ev := range events {
		if ev["type"] == "done" {
			t.Error("got done event on LLM error")
		}
	}
}

func TestChatHandler_MissingSession(t *testing.T) {
	store := session.NewInMemoryStore()
	mock := &mockLLM{}

	handler := ChatHandler(store, testScenarios(), mock, "test-model", discardLogger())

	body, _ := json.Marshal(map[string]string{
		"session_id": "nonexistent",
		"message":    "Hello",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/chat", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
	// SSE headers should NOT be set for a 404.
	if w.Header().Get("Content-Type") == "text/event-stream" {
		t.Error("SSE headers should not be set for a 404 response")
	}
}

func TestChatHandler_MalformedRequestBody(t *testing.T) {
	store := session.NewInMemoryStore()
	mock := &mockLLM{}

	handler := ChatHandler(store, testScenarios(), mock, "test-model", discardLogger())

	req := httptest.NewRequest(http.MethodPost, "/api/chat", strings.NewReader("not json"))
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
