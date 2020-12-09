package synthesizer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWord(t *testing.T) {
	assert.True(t, TaggedWord{Word: "olia"}.IsWord())
}

func TestWord_IsNot(t *testing.T) {
	assert.False(t, TaggedWord{Separator: ","}.IsWord())
	assert.False(t, TaggedWord{SentenceEnd: true}.IsWord())
}
