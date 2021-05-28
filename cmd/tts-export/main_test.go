package main

import (
	"flag"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeValue(t *testing.T) {
	tv := time.Time{}
	tm := timeValue{&tv}
	err := tm.Set("2021-05-28")
	assert.Nil(t, err)
	assert.Equal(t, time.Date(2021, 5, 28, 0, 0, 0, 0, time.UTC), tv)

	assert.NotNil(t, tm.Set("2021 05 28"))
	assert.NotNil(t, tm.Set("2021/05/28"))
}

func TestParseParams(t *testing.T) {
	p := &params{}
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	takeParams(fs, p)
	err := fs.Parse([]string{"-delete", "-to", "2021-05-28"})
	assert.Nil(t, err)
	assert.True(t, p.delete)
	assert.Equal(t, time.Date(2021, 5, 28, 0, 0, 0, 0, time.UTC), p.to)
}

func TestParseParams_Fail(t *testing.T) {
	params := &params{}
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	takeParams(fs, params)
	assert.NotNil(t, fs.Parse([]string{"-speakers", ""}))
	assert.NotNil(t, fs.Parse([]string{"-to", ""}))
	assert.NotNil(t, fs.Parse([]string{"-to", "2021-05--28"}))
}
