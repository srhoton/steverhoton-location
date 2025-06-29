package main

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetEnvVar(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		envValue     string
		defaultValue string
		expected     string
	}{
		{
			name:         "Environment variable exists",
			key:          "TEST_VAR",
			envValue:     "test_value",
			defaultValue: "default",
			expected:     "test_value",
		},
		{
			name:         "Environment variable does not exist",
			key:          "NON_EXISTENT_VAR",
			envValue:     "",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "Empty environment variable returns default",
			key:          "EMPTY_VAR",
			envValue:     "",
			defaultValue: "default",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variable if needed
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := getEnvVar(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInitializeHandler(t *testing.T) {
	ctx := context.Background()

	t.Run("Missing table name environment variable", func(t *testing.T) {
		// Ensure DYNAMODB_TABLE_NAME is not set
		os.Unsetenv("DYNAMODB_TABLE_NAME")

		handler, err := initializeHandler(ctx)
		assert.Error(t, err)
		assert.Nil(t, handler)
		assert.Contains(t, err.Error(), "DYNAMODB_TABLE_NAME environment variable is required")
	})

	t.Run("With table name set", func(t *testing.T) {
		// Set the required environment variable
		os.Setenv("DYNAMODB_TABLE_NAME", "test-table")
		defer os.Unsetenv("DYNAMODB_TABLE_NAME")

		// This test will fail in environments without AWS credentials,
		// which is expected in unit tests
		handler, err := initializeHandler(ctx)
		
		// We expect this to fail in test environment due to missing AWS credentials
		// In a real test, you would mock the AWS config loading
		if err != nil {
			assert.Contains(t, err.Error(), "failed to load AWS config")
		} else {
			require.NotNil(t, handler)
		}
	})
}