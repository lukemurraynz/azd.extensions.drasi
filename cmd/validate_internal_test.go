package cmd

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsCommandError(t *testing.T) {
	t.Parallel()

	assert.True(t, IsCommandError(&commandError{message: "failed", exitCode: 1}))
	assert.True(t, IsCommandError(fmt.Errorf("wrapped: %w", &commandError{message: "wrapped", exitCode: 2})))
	assert.False(t, IsCommandError(errors.New("plain error")))
}
