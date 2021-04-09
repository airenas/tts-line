package clean

import (
	"testing"

	"gotest.tools/assert"
)

func TestDropEmojis(t *testing.T) {
	assert.Equal(t, "", dropEmojis(""))
	assert.Equal(t, "a a", dropEmojis("a a"))
	assert.Equal(t, "a – a –", dropEmojis("a – a –"))
	assert.Equal(t, "a  a", dropEmojis("a 😀 a"))
	assert.Equal(t, "a  a", dropEmojis("a 😐 a"))
	assert.Equal(t, "a  a", dropEmojis("a 🐶 a"))
	assert.Equal(t, "a  a", dropEmojis("a 🏴󠁧󠁢󠁥󠁮󠁧󠁿 a"))
	assert.Equal(t, "a  a", dropEmojis("a 👏 a"))
	assert.Equal(t, "a b a", dropEmojis("a 👏👏b🏴󠁧󠁢󠁥󠁮󠁧󠁿a"))
}