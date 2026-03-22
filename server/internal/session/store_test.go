package session_test

import (
	"testing"

	"chatbot/server/internal/scenario"
	"chatbot/server/internal/session"
)

func TestInMemoryStore_SaveAndGet(t *testing.T) {
	store := session.NewInMemoryStore()
	cfg := scenario.ScenarioConfig{}

	sess, err := session.NewSession("sc1", cfg)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	if err := store.Save(sess); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := store.Get(sess.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID != sess.ID {
		t.Errorf("expected ID %q, got %q", sess.ID, got.ID)
	}
	if got.ScenarioID != "sc1" {
		t.Errorf("expected ScenarioID %q, got %q", "sc1", got.ScenarioID)
	}
}

func TestInMemoryStore_GetMissingReturnsError(t *testing.T) {
	store := session.NewInMemoryStore()

	_, err := store.Get("nonexistent-id")
	if err == nil {
		t.Error("expected error for missing session, got nil")
	}
}

func TestInMemoryStore_Delete(t *testing.T) {
	store := session.NewInMemoryStore()
	cfg := scenario.ScenarioConfig{}

	sess, _ := session.NewSession("sc2", cfg)
	_ = store.Save(sess)

	if err := store.Delete(sess.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := store.Get(sess.ID)
	if err == nil {
		t.Error("expected error after Delete, got nil")
	}
}

func TestInMemoryStore_DeleteMissingIsNoOp(t *testing.T) {
	store := session.NewInMemoryStore()

	// Deleting a non-existent session should not return an error.
	if err := store.Delete("no-such-id"); err != nil {
		t.Errorf("expected no error for Delete of missing session, got: %v", err)
	}
}

func TestInMemoryStore_SaveOverwrites(t *testing.T) {
	store := session.NewInMemoryStore()
	cfg := scenario.ScenarioConfig{}

	sess, _ := session.NewSession("sc3", cfg)
	_ = store.Save(sess)

	// Mutate and save again.
	sess.TurnCount = 42
	_ = store.Save(sess)

	got, err := store.Get(sess.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.TurnCount != 42 {
		t.Errorf("expected TurnCount 42 after overwrite, got %d", got.TurnCount)
	}
}

func TestInMemoryStore_ImplementsInterface(t *testing.T) {
	// Compile-time check that InMemoryStore satisfies SessionStore.
	var _ session.SessionStore = session.NewInMemoryStore()
}
