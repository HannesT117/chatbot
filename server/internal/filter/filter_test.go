package filter

import "testing"

func TestContainsBlocked(t *testing.T) {
	t.Run("case insensitive match", func(t *testing.T) {
		if !ContainsBlocked("Hello BOMB World", []string{"bomb"}) {
			t.Error("expected true for case-insensitive match")
		}
		if !ContainsBlocked("hello bomb world", []string{"BOMB"}) {
			t.Error("expected true for term in upper case")
		}
	})

	t.Run("multi-term blocklist", func(t *testing.T) {
		terms := []string{"bomb", "gun", "knife"}
		if !ContainsBlocked("I have a gun", terms) {
			t.Error("expected true when any term matches")
		}
		if ContainsBlocked("I have a pencil", terms) {
			t.Error("expected false when no term matches")
		}
	})

	t.Run("empty terms list", func(t *testing.T) {
		if ContainsBlocked("any text here", []string{}) {
			t.Error("expected false for empty blocklist")
		}
	})

	t.Run("empty text", func(t *testing.T) {
		if ContainsBlocked("", []string{"bomb"}) {
			t.Error("expected false for empty text")
		}
	})

	t.Run("partial match within word", func(t *testing.T) {
		// ContainsBlocked uses strings.Contains so partial matches are included.
		if !ContainsBlocked("bombing run", []string{"bomb"}) {
			t.Error("expected true for partial word match")
		}
	})
}

func TestContainsCanary(t *testing.T) {
	t.Run("exact match", func(t *testing.T) {
		if !ContainsCanary("prefix abc123token suffix", "abc123token") {
			t.Error("expected true for exact canary match")
		}
	})

	t.Run("not present", func(t *testing.T) {
		if ContainsCanary("some other text", "abc123token") {
			t.Error("expected false when canary not present")
		}
	})

	t.Run("partial substring does not match", func(t *testing.T) {
		// The canary check is exact substring; "abc123" is contained in "abc123token".
		if !ContainsCanary("abc123token", "abc123") {
			t.Error("expected true since abc123 is a substring of the text")
		}
		// But "abc123token" is not a substring of "abc123".
		if ContainsCanary("abc123", "abc123token") {
			t.Error("expected false when text is shorter than canary")
		}
	})

	t.Run("empty canary token", func(t *testing.T) {
		// strings.Contains always returns true for empty needle.
		if !ContainsCanary("anything", "") {
			t.Error("expected true for empty canary token (strings.Contains behaviour)")
		}
	})
}
