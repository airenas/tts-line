package ssml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsPause(t *testing.T) {
	assert.True(t, IsPause(&Pause{}))
	assert.False(t, IsPause(nil))
	assert.False(t, IsPause(&Text{}))
}
