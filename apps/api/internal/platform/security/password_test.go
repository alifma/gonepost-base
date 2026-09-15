package security

import "testing"

func TestHashAndVerify_CorrectPassword(t *testing.T) {
	hash, err := Hash("correct-horse-battery-staple")
	if err != nil {
		t.Fatalf("unexpected error hashing: %v", err)
	}

	ok, err := Verify("correct-horse-battery-staple", hash)
	if err != nil {
		t.Fatalf("unexpected error verifying: %v", err)
	}
	if !ok {
		t.Error("expected correct password to verify")
	}
}

func TestHashAndVerify_WrongPassword(t *testing.T) {
	hash, err := Hash("correct-horse-battery-staple")
	if err != nil {
		t.Fatalf("unexpected error hashing: %v", err)
	}

	ok, err := Verify("wrong-password", hash)
	if err != nil {
		t.Fatalf("unexpected error verifying: %v", err)
	}
	if ok {
		t.Error("expected wrong password to fail verification")
	}
}

func TestHash_SamePasswordDifferentHashes(t *testing.T) {
	hash1, err := Hash("same-password")
	if err != nil {
		t.Fatalf("unexpected error hashing: %v", err)
	}
	hash2, err := Hash("same-password")
	if err != nil {
		t.Fatalf("unexpected error hashing: %v", err)
	}

	if hash1 == hash2 {
		t.Error("expected two hashes of the same password to differ (random salt)")
	}
}
