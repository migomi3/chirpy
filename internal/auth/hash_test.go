package auth_test

import (
	"testing"

	"github.com/migomi3/internal/auth"
)

func TestHashPassword(t *testing.T) {
	password := "test"
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}
	if len(hash) != 60 || hash[:4] != "$2a$" {
		t.Fatalf("Hash %q is not in expected bcrypt format", hash)
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "test"
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	err = auth.CheckPasswordHash(password, hash)
	if err != nil {
		t.Fatalf(`CheckPasswordHash(%q, %q) returned error: %v`, password, hash, err)
	}
}
