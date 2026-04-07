package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetVersion(t *testing.T) {
	originalVersion := extensionVersion
	t.Cleanup(func() {
		extensionVersion = originalVersion
	})

	SetVersion("1.2.3-test")

	assert.Equal(t, "1.2.3-test", extensionVersion)
}
