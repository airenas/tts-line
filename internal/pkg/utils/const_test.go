package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestTypeStrin(t *testing.T) {
	assert.Equal(t, "RequestType:0", RequestTypeEnum(0).String())
	assert.Equal(t, "original", RequestOriginal.String())
	assert.Equal(t, "normalized", RequestNormalized.String())
	assert.Equal(t, "RequestType:100", RequestTypeEnum(100).String())
}
