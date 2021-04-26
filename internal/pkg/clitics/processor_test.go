package clitics

import (
	"testing"

	"github.com/airenas/tts-line/internal/pkg/clitics/service/api"

	"github.com/stretchr/testify/assert"
)

func TestProcess_None(t *testing.T) {
	pr, err := NewProcessor(map[string]bool{"bet": true}, &Phrases{})
	assert.Nil(t, err)
	assert.NotNil(t, pr)
	r, err := pr.Process([]*api.CliticsInput{{ID: 0, Type: "WORD", String: "bet"}})
	assert.Nil(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, 0, len(r))
}

func TestProcess_Finds(t *testing.T) {
	pr, err := NewProcessor(map[string]bool{"bet": true}, &Phrases{})
	assert.Nil(t, err)
	assert.NotNil(t, pr)
	r, err := pr.Process([]*api.CliticsInput{{ID: 0, Type: "WORD", String: "bet"},
		{ID: 1, Type: "SPACE", String: " "},
		{ID: 2, Type: "WORD", String: "kažkas", Mi: ""}})
	assert.Nil(t, err)
	assert.NotNil(t, r)
	if assert.Equal(t, 2, len(r)) {
		assert.Equal(t, 0, r[0].ID)
		assert.Equal(t, 2, r[1].ID)
	}
}

func TestProcess_Skips(t *testing.T) {
	pr, err := NewProcessor(map[string]bool{"bet": true}, &Phrases{})
	assert.Nil(t, err)
	assert.NotNil(t, pr)
	r, err := pr.Process([]*api.CliticsInput{{ID: 0, Type: "WORD", String: "bet"},
		{ID: 1, Type: "SEP", String: " "},
		{ID: 2, Type: "WORD", String: "kažkas", Mi: ""}})
	assert.Nil(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, 0, len(r))
}
