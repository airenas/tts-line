package clitics

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRead(t *testing.T) {
	cl, err := readClitics(strings.NewReader("bet\n   \nolia\n"))
	assert.Nil(t, err)
	assert.Equal(t, 2, len(cl))
}

func TestRead_FailEMpty(t *testing.T) {
	cl, err := readClitics(strings.NewReader(""))
	assert.NotNil(t, err)
	assert.Nil(t, cl)
}
