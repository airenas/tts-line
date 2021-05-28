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
	err := Export(Params{Exporter: expMock, Out: writer})
	assert.Nil(t, err)
	assert.Equal(t, "[]", writer.String())
}

func TestExport_Writes(t *testing.T) {
	initTest(t)
	writer := bytes.NewBufferString("")
	pegomock.When(expMock.All()).ThenReturn([]*mongodb.TextRecord{{ID: "1", Type: 1, Text: "olia"}}, nil)
	err := Export(Params{Exporter: expMock, Out: writer})
	assert.Nil(t, err)
	assert.Equal(t, "[{\"id\":\"1\",\"type\":1,\"text\":\"olia\",\"created\":\"0001-01-01T00:00:00Z\"}\n]", writer.String())
}

func TestExport_Sort(t *testing.T) {
	initTest(t)
	writer := bytes.NewBufferString("")
	tn := time.Time{}.Add(time.Second)
	pegomock.When(expMock.All()).ThenReturn([]*mongodb.TextRecord{{ID: "1", Type: 1, Text: "olia", Created: tn},
		{ID: "01", Type: 1, Text: "olia", Created: tn.Add(-time.Second)}}, nil)
	err := Export(Params{Exporter: expMock, Out: writer})
	assert.Nil(t, err)
	assert.Equal(t, "[{\"id\":\"01\",\"type\":1,\"text\":\"olia\",\"created\":\"0001-01-01T00:00:00Z\"}\n,{\"id\":\"1\",\"type\":1,\"text\":\"olia\",\"created\":\"0001-01-01T00:00:01Z\"}\n]",
		writer.String())
}

func TestExport_Fails(t *testing.T) {
	initTest(t)
	writer := bytes.NewBufferString("")
	pegomock.When(expMock.All()).ThenReturn(nil, errors.New("olia"))
	err := Export(Params{Exporter: expMock, Out: writer})
	assert.NotNil(t, err)
}

func TestExportFilter(t *testing.T) {
	initTest(t)
	writer := bytes.NewBufferString("")
	tn := time.Time{}.Add(time.Second)
	pegomock.When(expMock.All()).ThenReturn([]*mongodb.TextRecord{{ID: "1", Type: 1, Text: "olia", Created: tn},
		{ID: "01", Type: 1, Text: "olia", Created: tn.Add(time.Hour * 25)}}, nil)
	err := Export(Params{Exporter: expMock, Out: writer, Delete: true, To: tn.Add(time.Second)})
	assert.Nil(t, err)
	assert.Equal(t, "[{\"id\":\"1\",\"type\":1,\"text\":\"olia\",\"created\":\"0001-01-01T00:00:01Z\"}\n]", writer.String())
	p := expMock.VerifyWasCalledOnce().Delete(pegomock.AnyString()).GetCapturedArguments()
	assert.Equal(t, "1", p)
}

func TestExportDelete_Fails(t *testing.T) {
	initTest(t)
	writer := bytes.NewBufferString("")
	tn := time.Time{}.Add(time.Second)
	pegomock.When(expMock.All()).ThenReturn([]*mongodb.TextRecord{{ID: "1", Type: 1, Text: "olia", Created: tn},
		{ID: "01", Type: 1, Text: "olia", Created: tn.Add(-time.Second)}}, nil)
	pegomock.When(expMock.Delete(pegomock.AnyString())).ThenReturn(0, errors.New("olia"))
	err := Export(Params{Exporter: expMock, Out: writer, Delete: true})
	assert.NotNil(t, err)
}

func TestSortData(t *testing.T) {
	tn := time.Now()
	tests := []struct {
		d   []*mongodb.TextRecord
		pos []int
	}{
		{d: []*mongodb.TextRecord{{ID: "1", Type: 1}, {ID: "1", Type: 2}, {ID: "1", Type: 3}},
			pos: []int{0, 1, 2}},
		{d: []*mongodb.TextRecord{{ID: "1", Type: 2}, {ID: "1", Type: 3}, {ID: "1", Type: 1}},
			pos: []int{1, 2, 0}},
		{d: []*mongodb.TextRecord{{ID: "1", Type: 1, Created: tn}, {ID: "1", Type: 2, Created: tn.Add(time.Second * 5)},
			{ID: "2", Type: 1, Created: tn.Add(time.Second * 2)}},
			pos: []int{0, 1, 2}},
		{d: []*mongodb.TextRecord{{ID: "1", Type: 1, Created: tn}, {ID: "1", Type: 2, Created: tn.Add(time.Second * 5)},
			{ID: "2", Type: 1, Created: tn.Add(-time.Second * 2)}},
			pos: []int{1, 2, 0}},
		{d: []*mongodb.TextRecord{{ID: "1", Type: 1, Created: tn}, {ID: "1", Type: 2, Created: tn.Add(time.Second * 5)},
			{ID: "2", Type: 1, Created: tn.Add(-time.Second * 2)}, {ID: "2", Type: 3, Created: tn.Add(time.Second * 2)},
			{ID: "0", Type: 1, Created: tn.Add(time.Second * 2)}, {ID: "0", Type: 2, Created: tn.Add(time.Second * 2)}},
			pos: []int{2, 3, 0, 1, 4, 5}},
	}

	for _, tc := range tests {
		sd := sortData(tc.d)
		for i, v := range tc.pos {
			assert.Equal(t, tc.d[i], sd[v], "Fail case %d", i)
		}
	}
}

func TestFilterData(t *testing.T) {
	tn := time.Now()
	tests := []struct {
		d []*mongodb.TextRecord
		f time.Time
		c int
	}{
		{d: []*mongodb.TextRecord{{ID: "1", Type: 1, Created: tn}},
			c: 1},
		{d: []*mongodb.TextRecord{{ID: "1", Type: 1, Created: tn}},
			f: tn.Add(time.Second), c: 1},
		{d: []*mongodb.TextRecord{{ID: "1", Type: 1, Created: tn}},
			f: tn.Add(-time.Second), c: 0},
		{d: []*mongodb.TextRecord{{ID: "1", Type: 1, Created: tn}, {ID: "2", Type: 1, Created: tn.Add(time.Second * 2)}},
			f: tn.Add(time.Second), c: 1},
		{d: []*mongodb.TextRecord{{ID: "1", Type: 1, Created: tn}, {ID: "2", Type: 1, Created: tn.Add(time.Second * 2)},
			{ID: "1", Type: 2, Created: tn.Add(time.Second * 5)}},
			f: tn.Add(time.Second), c: 2},
		{d: []*mongodb.TextRecord{{ID: "1", Type: 3, Created: tn}},
			f: tn.Add(time.Second), c: 1},
	}

	for i, tc := range tests {
		fd := filterData(tc.d, tc.f)
		assert.Equal(t, tc.c, len(fd), "Fail case %d", i)
	}
}
