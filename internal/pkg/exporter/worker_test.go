package exporter

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"

	"github.com/airenas/tts-line/internal/pkg/mongodb"
	"github.com/airenas/tts-line/internal/pkg/test/mocks"
)

var (
	expMock *mocks.MockExporter
)

func initTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	expMock = mocks.NewMockExporter()
}

func TestExport(t *testing.T) {
	initTest(t)
	writer := bytes.NewBufferString("")
	pegomock.When(expMock.All()).ThenReturn([]*mongodb.TextRecord{}, nil)
	err := Export(expMock, writer)
	assert.Nil(t, err)
	assert.Equal(t, "[]\n", writer.String())
}

func TestExport_Writes(t *testing.T) {
	initTest(t)
	writer := bytes.NewBufferString("")
	pegomock.When(expMock.All()).ThenReturn([]*mongodb.TextRecord{{ID: "1", Type: 1, Text: "olia"}}, nil)
	err := Export(expMock, writer)
	assert.Nil(t, err)
	assert.Equal(t, "[{\"id\":\"1\",\"type\":1,\"text\":\"olia\",\"created\":\"0001-01-01T00:00:00Z\"}]\n", writer.String())
}

func TestExport_Sort(t *testing.T) {
	initTest(t)
	writer := bytes.NewBufferString("")
	pegomock.When(expMock.All()).ThenReturn([]*mongodb.TextRecord{{ID: "1", Type: 1, Text: "olia"}, {ID: "01", Type: 1, Text: "olia"}}, nil)
	err := Export(expMock, writer)
	assert.Nil(t, err)
	assert.Equal(t, "[{\"id\":\"01\",\"type\":1,\"text\":\"olia\",\"created\":\"0001-01-01T00:00:00Z\"},{\"id\":\"1\",\"type\":1,\"text\":\"olia\",\"created\":\"0001-01-01T00:00:00Z\"}]\n",
		writer.String())
}

func TestExport_Fails(t *testing.T) {
	initTest(t)
	writer := bytes.NewBufferString("")
	pegomock.When(expMock.All()).ThenReturn(nil, errors.New("olia"))
	err := Export(expMock, writer)
	assert.NotNil(t, err)
}

func TestCompare(t *testing.T) {
	tests := []struct {
		i1 mongodb.TextRecord
		i2 mongodb.TextRecord
		v  bool
	}{
		{i1: mongodb.TextRecord{ID: "1", Type: 1}, i2: mongodb.TextRecord{ID: "2", Type: 1}, v: true},
		{i1: mongodb.TextRecord{ID: "3", Type: 1}, i2: mongodb.TextRecord{ID: "2", Type: 1}, v: false},
		{i1: mongodb.TextRecord{ID: "1", Type: 1}, i2: mongodb.TextRecord{ID: "1", Type: 2}, v: true},
		{i1: mongodb.TextRecord{ID: "1", Type: 2}, i2: mongodb.TextRecord{ID: "1", Type: 1}, v: false},
		{i1: mongodb.TextRecord{ID: "1", Type: 1, Created: time.Now()},
			i2: mongodb.TextRecord{ID: "1", Type: 1, Created: time.Now().Add(time.Second)}, v: true},
		{i1: mongodb.TextRecord{ID: "1", Type: 1, Created: time.Now()},
			i2: mongodb.TextRecord{ID: "1", Type: 1, Created: time.Now().Add(-time.Second)}, v: false},
	}

	for _, tc := range tests {
		v := compare(&tc.i1, &tc.i2)
		assert.Equal(t, tc.v, v)
	}
}
