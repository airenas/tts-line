package clitics

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPhrasesRead(t *testing.T) {
	ph, err := readPhrases(strings.NewReader("bet ko9ks:l\nkuri4s:l ne kuri4s:l\nolia olia:l\n//comment\n"))
	assert.Nil(t, err)
	assert.Equal(t, 2, len(ph.wordMap))
	assert.Equal(t, 1, len(ph.lemmaMap))
}

func TestPhrasesRead_FailEmpty(t *testing.T) {
	cl, err := readClitics(strings.NewReader(""))
	assert.NotNil(t, err)
	assert.Nil(t, cl)
}

func TestPhrasesReadLine(t *testing.T) {
	ph, err := readLine("bet ko9ks:l")
	assert.Nil(t, err)
	if assert.Equal(t, 2, len(ph)) {
		assert.Equal(t, &phrase{word:"bet"}, ph[0])
		assert.Equal(t, &phrase{lemma:"koks", accent: 202}, ph[1])
	}
	_, err = readLine("bet")
	assert.NotNil(t, err)
	ph, err = readLine("ko9ks bet ko9ks:l")
	assert.Nil(t, err)
	if assert.Equal(t, 3, len(ph)) {
		assert.Equal(t, &phrase{word:"koks", accent: 202}, ph[0])
		assert.Equal(t, &phrase{word:"bet"}, ph[1])
		assert.Equal(t, &phrase{lemma:"koks", accent: 202}, ph[2])
	}
}
