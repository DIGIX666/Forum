package script

import (
	"testing"
)

func TestGenerateHashAndComparePassword(t *testing.T) {
	password := "motdepasse"
	hash := GenerateHash(password)
	if hash == "" {
		t.Fatal("The hash must not be empty")
	}
	if !ComparePassword(hash, password) {
		t.Error("The password and the hash should be the same")
	}
}

func TestGenerateRandomString(t *testing.T) {
	s := GenerateRandomString()
	if len(s) != 10 {
		t.Errorf("The generated string must be 10 characters long, obtain: %d", len(s))
	}
}

func TestGeneratePostID(t *testing.T) {
	id := GeneratePostID()
	if len(id) != 7 {
		t.Errorf("The post ID must be 7 characters long, obtained from: %d", len(id))
	}
}

func TestGenerateCommentID(t *testing.T) {
	id := GenerateCommentID()
	if len(id) != 7 {
		t.Errorf("The comment ID must be 7 characters long, obtained from: %d", len(id))
	}
}
