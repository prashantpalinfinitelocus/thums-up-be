package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateOTP(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"6 digit OTP", 6},
		{"4 digit OTP", 4},
		{"8 digit OTP", 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			otp, err := GenerateOTP(tt.length)

			assert.NoError(t, err)
			assert.Len(t, otp, tt.length)

			// Verify all characters are digits
			for _, char := range otp {
				assert.True(t, char >= '0' && char <= '9', "OTP should only contain digits")
			}
		})
	}
}

func TestGenerateReferralCode(t *testing.T) {
	// Generate multiple codes to ensure uniqueness
	codes := make(map[string]bool)
	iterations := 100

	for i := 0; i < iterations; i++ {
		code, err := GenerateReferralCode()

		assert.NoError(t, err)
		assert.Len(t, code, 8, "Referral code should be 8 characters")

		// Verify all characters are alphanumeric
		for _, char := range code {
			isValid := (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')
			assert.True(t, isValid, "Referral code should only contain A-Z and 0-9")
		}

		// Check uniqueness
		assert.False(t, codes[code], "Referral code should be unique")
		codes[code] = true
	}
}

func TestFormatPhoneNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"10 digit number", "9876543210", "919876543210"},
		{"With country code", "919876543210", "919876543210"},
		{"With spaces", " 9876543210 ", "919876543210"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatPhoneNumber(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected bool
	}{
		{"Valid email", "test@example.com", true},
		{"Valid email with subdomain", "user@mail.example.com", true},
		{"Invalid - no @", "testexample.com", false},
		{"Invalid - no domain", "test@", false},
		{"Invalid - no TLD", "test@example", false},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidEmail(tt.email)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidPhoneNumber(t *testing.T) {
	tests := []struct {
		name     string
		phone    string
		expected bool
	}{
		{"Valid 10 digit", "9876543210", true},
		{"Invalid - 9 digits", "987654321", false},
		{"Invalid - 11 digits", "98765432101", false},
		{"Invalid - with letters", "987654321a", false},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidPhoneNumber(tt.phone)
			assert.Equal(t, tt.expected, result)
		})
	}
}
