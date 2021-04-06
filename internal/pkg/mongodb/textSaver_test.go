package mongodb

import (
	"testing"

	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestNewTextSaver(t *testing.T) {
	tpr, _ := NewSessionProvider("mongo")
	pr, err := NewTextSaver(tpr)
	assert.NotNil(t, pr)
	assert.Nil(t, err)
}

func TestToRecord(t *testing.T) {
	res := toRecord("ii", "text", utils.RequestOriginal)
	assert.Equal(t, "ii", res.ID)
	assert.Equal(t, "text", res.Text)
	assert.Equal(t, 1, res.Type)
}
