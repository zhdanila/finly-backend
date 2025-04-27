package validator

import (
	"testing"
)

func TestValidateCommaArray(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Valid integers",
			input:    "1,2,3,4",
			expected: true,
		},
		{
			name:     "Invalid integer",
			input:    "1,2,abc,4",
			expected: false,
		},
	}

	v := NewValidator()
	_ = v.Register("comma_array", ValidateCommaArray)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.validator.Var(tt.input, "comma_array")
			if (err == nil) != tt.expected {
				t.Errorf("ValidateCommaArray() = %v, want %v", err == nil, tt.expected)
			}
		})
	}
}

func TestValidateUUID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Valid UUID",
			input:    "123e4567-e89b-12d3-a456-426614174000",
			expected: true,
		},
		{
			name:     "Invalid UUID",
			input:    "invalid-uuid",
			expected: false,
		},
	}

	v := NewValidator()
	_ = v.Register("uuid", ValidateUUID)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.validator.Var(tt.input, "uuid")
			if (err == nil) != tt.expected {
				t.Errorf("ValidateUUID() = %v, want %v", err == nil, tt.expected)
			}
		})
	}
}

func TestValidateTimestamp(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Valid RFC3339 timestamp",
			input:    "2025-04-27T15:04:05Z",
			expected: true,
		},
		{
			name:     "Invalid RFC3339 timestamp",
			input:    "2025-04-27 15:04:05",
			expected: false,
		},
	}

	v := NewValidator()
	_ = v.Register("timestamp", ValidateTimestamp)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.validator.Var(tt.input, "timestamp")
			if (err == nil) != tt.expected {
				t.Errorf("ValidateTimestamp() = %v, want %v", err == nil, tt.expected)
			}
		})
	}
}
