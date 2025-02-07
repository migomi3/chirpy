package auth_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/migomi3/internal/auth"
)

func TestValidateJWT(t *testing.T) {
	userID := uuid.New()
	validToken, _ := auth.MakeJWT(userID, "secret", time.Hour)

	tests := []struct {
		name        string
		tokenString string
		tokenSecret string
		wantUserID  uuid.UUID
		wantErr     bool
	}{
		{
			name:        "Valid token",
			tokenString: validToken,
			tokenSecret: "secret",
			wantUserID:  userID,
			wantErr:     false,
		},
		{
			name:        "Invalid token",
			tokenString: "invalid.token.string",
			tokenSecret: "secret",
			wantUserID:  uuid.Nil,
			wantErr:     true,
		},
		{
			name:        "Wrong secret",
			tokenString: validToken,
			tokenSecret: "wrong_secret",
			wantUserID:  uuid.Nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUserID, err := auth.ValidateJWT(tt.tokenString, tt.tokenSecret)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateJWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotUserID != tt.wantUserID {
				t.Errorf("ValidateJWT() gotUserID = %v, want %v", gotUserID, tt.wantUserID)
			}
		})
	}
}

func TestGetBearerToken(t *testing.T) {
	header := http.Header{}

	tests := []struct {
		input          []string
		expectedOutput string
	}{
		{[]string{"Authorization", "Bearer testing"}, "testing"},
		{[]string{"Authorization", "Bearer   gaerighuhi  "}, "gaerighuhi"},
	}

	for _, testCase := range tests {
		header.Set(testCase.input[0], testCase.input[1])
		result, err := auth.GetBearerToken(header)
		if err != nil {
			t.Fatalf("Error in GetBearerToken: %v\n", err)
		}

		if result != testCase.expectedOutput {
			t.Fatalf("Result [%q] does not match expected output [%q]\n", result, testCase.expectedOutput)
		}
	}
}
