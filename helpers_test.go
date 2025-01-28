package main

import (
	"fmt"
	"testing"
)

func TestCleanMessage(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"test", "test"},
		{"Test", "Test"},
		{"kerfuffle", "****"},
		{"fucking kerfuffle", "fucking ****"},
		{"testing Sharbert", "testing ****"},
		{"uppercase Fornax", "uppercase ****"},
		{"Fornax!", "Fornax!"},
		{"sharbert, ", "sharbert, "},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Running Input: %s\n", tc.input), func(t *testing.T) {
			result := cleanMessage(tc.input)

			if len(tc.expected) != len(result) {
				t.Errorf("lengths don't match: '%v' vs '%v'", result, tc.expected)
			}

			for i := range tc.expected {
				if tc.expected[i] != result[i] {
					t.Errorf("Expected output [%s] does not match Actual output [%s]", tc.expected, result)
				}
			}
		})
	}
}
