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

func TestFloatEqual(t *testing.T) {
	assert.True(t, FloatEquals(0.0, 0.0))
	assert.True(t, FloatEquals(-0, 0))
	assert.True(t, FloatEquals(0.00001, 0.00001))
	assert.True(t, FloatEquals(1000000000, 1000000000))
	assert.True(t, FloatEquals(2.0/3, 4.0/6))
	assert.False(t, FloatEquals(200.0/3000, 200.0/3001))
}
