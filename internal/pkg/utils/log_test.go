package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrimString(t *testing.T) {
	assert.Equal(t, "olia", trimString("olia", 10))
	assert.Equal(t, "0123456789", trimString("0123456789", 10))
	assert.Equal(t, "0123456789...", trimString("01234567890", 10))
}
