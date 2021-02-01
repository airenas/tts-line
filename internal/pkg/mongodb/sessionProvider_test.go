package mongodb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSessionProvider(t *testing.T) {
	pr, err := NewSessionProvider("mongo://url")
	assert.NotNil(t, pr)
	assert.Nil(t, err)
}
