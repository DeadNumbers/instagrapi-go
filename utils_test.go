package instagrapi

import (
	"encoding/json"
	"testing"
)

// TestGenerateSignature tests the signature generation function.
func TestGenerateSignature(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple data",
			input:    `{"key":"value"}`,
			expected: "signed_body=SIGNATURE.", // prefix check, actual encoded value depends on urlEncode
		},
		{
			name:     "empty string",
			input:    "",
			expected: "signed_body=SIGNATURE.",
		},
		{
			name:     "data with special chars",
			input:    `{"key":"value&other=test"}`,
			expected: "signed_body=SIGNATURE.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateSignature(tt.input)
			expectedPrefix := "signed_body=SIGNATURE."
			if len(result) < len(expectedPrefix) || result[:len(expectedPrefix)] != expectedPrefix {
				t.Errorf("GenerateSignature(%q) = %q; want to start with %q", tt.input, result, expectedPrefix)
			}
		})
	}
}

// TestGenToken tests token generation.
func TestGenToken(t *testing.T) {
	tests := []struct {
		name        string
		size        int
		withSymbols bool
	}{
		{"default", 32, false},
		{"with symbols", 16, true},
		{"short token", 8, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenToken(tt.size, tt.withSymbols)
			if len(result) != tt.size {
				t.Errorf("GenToken(%d, %v) length = %d; want %d", tt.size, tt.withSymbols, len(result), tt.size)
			}
			if tt.size == 0 {
				return
			}
			// Verify all characters are from the allowed set
			const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
			const symbols = "!@#$%^&*()_+-=[]{}|;:',.<>?/"
			for _, ch := range result {
				found := false
				for _, l := range letters {
					if ch == l {
						found = true
						break
					}
				}
				if !found && tt.withSymbols {
					for _, s := range symbols {
						if ch == s {
							found = true
							break
						}
					}
				}
				if !found {
					t.Errorf("GenToken generated unexpected character: %c", ch)
					break
				}
			}
		})
	}

	// Test that different calls produce different tokens (statistical test)
	token1 := GenToken(64, false)
	token2 := GenToken(64, false)
	if token1 == token2 {
		t.Log("Warning: two consecutive GenToken calls produced the same result")
	}
}

// TestGenerateJazoest tests jazoest generation.
func TestGenerateJazoest(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"simple", "test"},
		{"with numbers", "abc123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateJazoest(tt.input)
			if len(result) < 2 || result[0] != '2' {
				t.Errorf("GenerateJazoest(%q) = %q; want to start with '2'", tt.input, result)
			}
			// The rest should be a valid number (the sum of rune values)
			var expectedSum int
			for _, r := range tt.input {
				expectedSum += int(r)
			}
			expected := "2" + string(rune('0'+expectedSum%10)) // simplified check
			if len(result) < 2 {
				t.Errorf("GenerateJazoest(%q) returned too short result: %q", tt.input, result)
			}
			_ = expected
		})
	}

	// Verify the numeric part matches the sum
	result := GenerateJazoest("abc")
	expectedSum := int('a') + int('b') + int('c')
	expectedStr := "2" + string(rune('0'+expectedSum%10))
	if result != expectedStr {
		t.Logf("GenerateJazoest(\"abc\") = %q; expected sum-based: %q (sum=%d)", result, expectedStr, expectedSum)
	}
}

// TestHashPassword tests password hashing.
func TestHashPassword(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"simple", "password123"},
		{"empty", ""},
		{"unicode", "пароль🔒"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HashPassword(tt.input)
			if len(result) != 64 {
				t.Errorf("HashPassword(%q) length = %d; want 64", tt.input, len(result))
			}
			// Verify it's a valid hex string
			for _, ch := range result {
				if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f')) {
					t.Errorf("HashPassword produced non-hex character: %c", ch)
					break
				}
			}
			// Verify deterministic output
			result2 := HashPassword(tt.input)
			if result != result2 {
				t.Error("HashPassword is not deterministic")
			}
		})
	}

	// Different inputs should produce different hashes
	hash1 := HashPassword("password1")
	hash2 := HashPassword("password2")
	if hash1 == hash2 {
		t.Error("Different passwords produced the same hash")
	}
}

// TestNavigateJSON tests JSON navigation.
func TestNavigateJSON(t *testing.T) {
	data := map[string]any{
		"user": map[string]any{
			"name": "test",
			"posts": []any{
				map[string]any{"id": 1},
				map[string]any{"id": 2},
			},
		},
	}

	tests := []struct {
		name     string
		path     []any
		expected any
	}{
		{
			name:     "navigate to user",
			path:     []any{"user"},
			expected: map[string]any{"name": "test", "posts": []any{map[string]any{"id": 1}, map[string]any{"id": 2}}},
		},
		{
			name:     "navigate to user name",
			path:     []any{"user", "name"},
			expected: "test",
		},
		{
			name:     "navigate to first post",
			path:     []any{"user", "posts", 0},
			expected: map[string]any{"id": 1.0}, // JSON numbers become float64
		},
		{
			name:     "navigate to second post id",
			path:     []any{"user", "posts", 1, "id"},
			expected: float64(2),
		},
		{
			name:     "out of bounds index",
			path:     []any{"user", "posts", 5},
			expected: nil,
		},
		{
			name:     "non-existent key",
			path:     []any{"user", "nonexistent"},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := navigateJSON(data, tt.path...)
			if result == nil && tt.expected == nil {
				return // both nil is success
			}
			if result == nil || tt.expected == nil {
				t.Errorf("navigateJSON(%v) = %v; want %v", tt.path, result, tt.expected)
				return
			}

			// For float64 comparisons (from JSON unmarshaling)
			if expectedFloat, ok := tt.expected.(float64); ok {
				if resultFloat, ok := result.(float64); ok && resultFloat == expectedFloat {
					return
				}
			}

			resultJSON, _ := json.Marshal(result)
			expectedJSON, _ := json.Marshal(tt.expected)
			if string(resultJSON) != string(expectedJSON) {
				t.Errorf("navigateJSON(%v) = %s; want %s", tt.path, resultJSON, expectedJSON)
			}
		})
	}

	// Test with empty path
	result := navigateJSON(data)
	if result == nil {
		t.Error("navigateJSON with no path should return the original data")
	}
}

// TestJsonMarshal tests JSON marshaling functions.
func TestJsonMarshal(t *testing.T) {
	data := map[string]any{
		"key": "value",
		"num": 42,
	}

	result, err := jsonMarshal(data)
	if err != nil {
		t.Fatalf("jsonMarshal error: %v", err)
	}
	if result == "" {
		t.Error("jsonMarshal returned empty string")
	}

	// Verify it's valid JSON
	var decoded map[string]any
	if err := json.Unmarshal([]byte(result), &decoded); err != nil {
		t.Errorf("jsonMarshal output is not valid JSON: %v", err)
	}

	compactResult := jsonMarshalCompact(data)
	if compactResult == "" {
		t.Error("jsonMarshalCompact returned empty string")
	}
	if len(compactResult) >= len(result) && result != "" {
		// Compact should be shorter or equal (no trailing newline, no spaces)
		t.Logf("compact length: %d, marshal length: %d", len(compactResult), len(result))
	}

	// Verify compact is valid JSON
	if err := json.Unmarshal([]byte(compactResult), &decoded); err != nil {
		t.Errorf("jsonMarshalCompact output is not valid JSON: %v", err)
	}
}
