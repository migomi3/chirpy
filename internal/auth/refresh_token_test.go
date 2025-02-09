package auth_test

import (
	"testing"

	"github.com/migomi3/internal/auth"
)

func TestMakeRefreshToken(t *testing.T) {
	t.Run("Is Valid", func(t *testing.T) {
		tokenString, err := auth.MakeRefreshToken()
		if err != nil {
			t.Error(err)
		}

		if tokenString == "" {
			t.Error("token is empty")
		}

		if len(tokenString) != 64 {
			t.Errorf("token has invalid length: %d", len(tokenString))
		}
	})
}
