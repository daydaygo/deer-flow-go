package channels

import (
	"testing"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"hello", ""},
		{"/help", "help"},
		{"/new", "new"},
		{"/status", "status"},
		{"/new conversation", "new"},
		{"/help me please", "help"},
	}

	for _, tt := range tests {
		result := parseCommand(tt.input)
		if result != tt.expected {
			t.Errorf("parseCommand(%s) = %s, expected %s", tt.input, result, tt.expected)
		}
	}
}
