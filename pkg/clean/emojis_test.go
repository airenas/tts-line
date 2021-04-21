package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDropEmojis(t *testing.T) {
	assert.Equal(t, "", DropEmojis(""))
	assert.Equal(t, "a a", DropEmojis("a a"))
	assert.Equal(t, "a – a –", DropEmojis("a – a –"))
	assert.Equal(t, "a  a", DropEmojis("a 😀 a"))
	assert.Equal(t, "a  a", DropEmojis("a 😐 a"))
	assert.Equal(t, "a  a", DropEmojis("a 🐶 a"))
	assert.Equal(t, "a  a", DropEmojis("a 🏴󠁧󠁢󠁥󠁮󠁧󠁿 a"))
	assert.Equal(t, "a  a", DropEmojis("a 👏 a"))
	assert.Equal(t, "a b a", DropEmojis("a 👏👏b🏴󠁧󠁢󠁥󠁮󠁧󠁿a"))
}
