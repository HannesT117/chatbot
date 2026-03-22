package session_test

import (
	"testing"

	"chatbot/server/internal/scenario"
	"chatbot/server/internal/session"
)

func newTestSession(t *testing.T) *session.Session {
	t.Helper()
	cfg := scenario.ScenarioConfig{MaxTurns: 10, TokenBudget: 4096}
	s, err := session.NewSession("test-scenario", cfg)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	return s
}

func TestNewSession_FieldsInitialised(t *testing.T) {
	s := newTestSession(t)

	if s.ID == "" {
		t.Error("expected non-empty ID")
	}
	if s.CanaryToken == "" {
		t.Error("expected non-empty CanaryToken")
	}
	if len(s.CanaryToken) != 32 { // 16 bytes → 32 hex chars
		t.Errorf("expected CanaryToken length 32, got %d", len(s.CanaryToken))
	}
	if s.ScenarioID != "test-scenario" {
		t.Errorf("expected ScenarioID %q, got %q", "test-scenario", s.ScenarioID)
	}
	if s.TurnCount != 0 {
		t.Errorf("expected TurnCount 0, got %d", s.TurnCount)
	}
	if len(s.Messages) != 0 {
		t.Errorf("expected empty Messages, got %d", len(s.Messages))
	}
	if s.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
}

func TestNewSession_UniqueIDs(t *testing.T) {
	cfg := scenario.ScenarioConfig{}
	s1, _ := session.NewSession("sc", cfg)
	s2, _ := session.NewSession("sc", cfg)

	if s1.ID == s2.ID {
		t.Error("expected unique session IDs, got duplicates")
	}
	if s1.CanaryToken == s2.CanaryToken {
		t.Error("expected unique canary tokens, got duplicates")
	}
}

func TestTurnLimitReached(t *testing.T) {
	s := newTestSession(t)

	if s.TurnLimitReached(5) {
		t.Error("expected false for TurnCount=0 < maxTurns=5")
	}

	s.AddTurn("hello", "hi")
	s.AddTurn("world", "there")

	if s.TurnLimitReached(5) {
		t.Error("expected false for TurnCount=2 < maxTurns=5")
	}

	// Fill up to the limit
	s.AddTurn("a", "b")
	s.AddTurn("c", "d")
	s.AddTurn("e", "f")

	if !s.TurnLimitReached(5) {
		t.Error("expected true for TurnCount=5 >= maxTurns=5")
	}

	if !s.TurnLimitReached(3) {
		t.Error("expected true for TurnCount=5 >= maxTurns=3")
	}
}

func TestAddTurn_AppendsTwoMessages(t *testing.T) {
	s := newTestSession(t)

	s.AddTurn("user message", "assistant reply")

	if len(s.Messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(s.Messages))
	}
	if s.TurnCount != 1 {
		t.Errorf("expected TurnCount 1, got %d", s.TurnCount)
	}

	s.AddTurn("second user", "second assistant")

	if len(s.Messages) != 4 {
		t.Errorf("expected 4 messages, got %d", len(s.Messages))
	}
	if s.TurnCount != 2 {
		t.Errorf("expected TurnCount 2, got %d", s.TurnCount)
	}
}

func TestApplySlidingWindow_NoOpWhenUnderBudget(t *testing.T) {
	s := newTestSession(t)
	s.AddTurn("hi", "hello") // 2+5 = 7 chars → ~1 token

	s.ApplySlidingWindow(1000)

	if len(s.Messages) != 2 {
		t.Errorf("expected 2 messages to remain, got %d", len(s.Messages))
	}
}

func TestApplySlidingWindow_DropsOldestPairs(t *testing.T) {
	s := newTestSession(t)

	// Add 3 turns with substantial content so we can trigger the window.
	// Each turn: user ~100 chars, assistant ~100 chars → ~200 chars per turn → ~50 tokens.
	user100 := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" // 98 chars
	asst100 := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"  // 98 chars

	s.AddTurn(user100, asst100) // turn 1
	s.AddTurn(user100, asst100) // turn 2
	s.AddTurn(user100, asst100) // turn 3
	// total: 6 messages, ~588 chars / 4 = ~147 tokens

	// Budget of 80 tokens should cause the oldest pair(s) to be dropped.
	s.ApplySlidingWindow(80)

	if len(s.Messages) < 2 {
		t.Errorf("expected at least 2 messages (last turn retained), got %d", len(s.Messages))
	}

	// After trimming, estimate should be within budget or we're at the minimum (2 msgs).
	if len(s.Messages) == 2 {
		// Minimum reached — that's fine.
		return
	}
	remaining := len(s.Messages) / 2 // number of turns remaining
	if remaining >= 3 {
		t.Errorf("expected fewer than 3 turns after sliding window, got %d", remaining)
	}
}

func TestApplySlidingWindow_KeepsLastTurn(t *testing.T) {
	s := newTestSession(t)

	// Very large single turn — should still keep it even if over budget.
	bigMsg := string(make([]byte, 10000)) // 10000 chars → 2500 tokens
	s.AddTurn(bigMsg, bigMsg)

	s.ApplySlidingWindow(10) // tiny budget

	// Must always keep at least the last turn (2 messages).
	if len(s.Messages) != 2 {
		t.Errorf("expected 2 messages retained (minimum), got %d", len(s.Messages))
	}
}
