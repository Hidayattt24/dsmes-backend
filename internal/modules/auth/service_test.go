package auth

import "testing"

func TestHashRefreshToken(t *testing.T) {
	// Must be deterministic for the same input (lookup key stability).
	a := HashRefreshToken("some-random-token")
	b := HashRefreshToken("some-random-token")
	if a != b {
		t.Error("expected HashRefreshToken to be deterministic")
	}

	// Must be a SHA-256 hex digest (64 chars) so it fits the column.
	if len(a) != 64 {
		t.Errorf("expected 64-char hex digest, got %d chars: %s", len(a), a)
	}

	// Different inputs must produce different digests.
	c := HashRefreshToken("some-other-token")
	if a == c {
		t.Error("expected different tokens to produce different hashes")
	}
}
