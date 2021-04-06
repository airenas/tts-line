package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestURL(t *testing.T) {
	_, err := checkURL("")
	assert.NotNil(t, err)
	_, err = checkURL(":8080")
	assert.NotNil(t, err)
	_, err = checkURL("http://")
	assert.NotNil(t, err)

	res, err := checkURL("http://local:8080")
	assert.Nil(t, err)

	assert.Equal(t, "http://local:8080", res)
}
