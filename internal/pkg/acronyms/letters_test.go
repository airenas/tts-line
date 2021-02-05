package acronyms

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReturns(t *testing.T) {
	prv, err := NewLetters()
	assert.Nil(t, err)
	res, err := prv.Process("a", "1")
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 1, len(res))
	if len(res) == 1 {
		assert.Equal(t, "a", res[0].Word)
		assert.Equal(t, "ą", res[0].WordTrans)
		assert.Equal(t, "ą", res[0].Syll)
		assert.Equal(t, "ą3", res[0].UserTrans)
	}
}

func TestReturnsLower(t *testing.T) {
	prv, err := NewLetters()
	assert.Nil(t, err)
	res, err := prv.Process("B", "1")
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 1, len(res))
	if len(res) == 1 {
		assert.Equal(t, "b", res[0].Word)
		assert.Equal(t, "bė", res[0].WordTrans)
		assert.Equal(t, "bė", res[0].Syll)
		assert.Equal(t, "bė3", res[0].UserTrans)
	}
}

func TestReturnsSeveral(t *testing.T) {
	prv, err := NewLetters()
	assert.Nil(t, err)
	res, err := prv.Process("BA", "1")
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 1, len(res))
	if len(res) == 1 {
		assert.Equal(t, "ba", res[0].Word)
		assert.Equal(t, "bėą", res[0].WordTrans)
		assert.Equal(t, "bė-ą", res[0].Syll)
		assert.Equal(t, "bėą3", res[0].UserTrans)
	}
}

func TestAccent(t *testing.T) {
	prv, err := NewLetters()
	assert.Nil(t, err)
	res, err := prv.Process("AABA", "1")
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 1, len(res))
	if len(res) == 1 {
		assert.Equal(t, "aaba", res[0].Word)
		assert.Equal(t, "ąąbėą", res[0].WordTrans)
		assert.Equal(t, "ą-ą-bė-ą", res[0].Syll)
		assert.Equal(t, "ąąbėą3", res[0].UserTrans)
	}
}

func TestNewWord(t *testing.T) {
	prv, err := NewLetters()
	assert.Nil(t, err)
	res, err := prv.Process("AAWBA", "1")
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 3, len(res))
	if len(res) == 3 {
		assert.Equal(t, "aa", res[0].Word)
		assert.Equal(t, "ąą", res[0].WordTrans)
		assert.Equal(t, "ą-ą", res[0].Syll)
		assert.Equal(t, "ąą3", res[0].UserTrans)

		assert.Equal(t, "w", res[1].Word)
		assert.Equal(t, "dablvė", res[1].WordTrans)
		assert.Equal(t, "dabl-vė", res[1].Syll)
		assert.Equal(t, "da4blvė", res[1].UserTrans)

		assert.Equal(t, "ba", res[2].Word)
		assert.Equal(t, "bėą", res[2].WordTrans)
		assert.Equal(t, "bė-ą", res[2].Syll)
		assert.Equal(t, "bėą3", res[2].UserTrans)
	}
}

func TestNewWordFirst(t *testing.T) {
	prv, err := NewLetters()
	assert.Nil(t, err)
	res, err := prv.Process("WBA", "1")
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 2, len(res))
	if len(res) == 2 {
		assert.Equal(t, "w", res[0].Word)
		assert.Equal(t, "dablvė", res[0].WordTrans)
		assert.Equal(t, "dabl-vė", res[0].Syll)
		assert.Equal(t, "da4blvė", res[0].UserTrans)

		assert.Equal(t, "ba", res[1].Word)
		assert.Equal(t, "bėą", res[1].WordTrans)
		assert.Equal(t, "bė-ą", res[1].Syll)
		assert.Equal(t, "bėą3", res[1].UserTrans)
	}
}

func TestNewWordLast(t *testing.T) {
	prv, err := NewLetters()
	assert.Nil(t, err)
	res, err := prv.Process("AAW", "1")
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 2, len(res))
	if len(res) == 2 {
		assert.Equal(t, "aa", res[0].Word)
		assert.Equal(t, "ąą", res[0].WordTrans)
		assert.Equal(t, "ą-ą", res[0].Syll)
		assert.Equal(t, "ąą3", res[0].UserTrans)

		assert.Equal(t, "w", res[1].Word)
		assert.Equal(t, "dablvė", res[1].WordTrans)
		assert.Equal(t, "dabl-vė", res[1].Syll)
		assert.Equal(t, "da4blvė", res[1].UserTrans)
	}
}

func TestIgnoreDot(t *testing.T) {
	prv, err := NewLetters()
	assert.Nil(t, err)
	res, err := prv.Process("A.", "1")
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 1, len(res))
	if len(res) == 1 {
		assert.Equal(t, "a", res[0].Word)
		assert.Equal(t, "ą", res[0].WordTrans)
		assert.Equal(t, "ą", res[0].Syll)
		assert.Equal(t, "ą3", res[0].UserTrans)
	}
}

func TestNoFail(t *testing.T) {
	prv, err := NewLetters()
	assert.Nil(t, err)
	res, err := prv.Process("'.A.", "1")
	assert.Nil(t, err)
	assert.NotNil(t, res)
	if assert.Equal(t, 1, len(res)) {
		assert.Equal(t, "a", res[0].Word)
		assert.Equal(t, "ą", res[0].WordTrans)
		assert.Equal(t, "ą", res[0].Syll)
		assert.Equal(t, "ą3", res[0].UserTrans)
	}
}

func TestNoFailSeveralWords(t *testing.T) {
	prv, err := NewLetters()
	assert.Nil(t, err)
	res, err := prv.Process("'.A.W>", "1")
	assert.Nil(t, err)
	assert.NotNil(t, res)
	if assert.Equal(t, 2, len(res)) {
		assert.Equal(t, "a", res[0].Word)
		assert.Equal(t, "ą", res[0].WordTrans)
		assert.Equal(t, "ą", res[0].Syll)
		assert.Equal(t, "ą3", res[0].UserTrans)

		assert.Equal(t, "w", res[1].Word)
		assert.Equal(t, "dablvė", res[1].WordTrans)
		assert.Equal(t, "dabl-vė", res[1].Syll)
		assert.Equal(t, "da4blvė", res[1].UserTrans)
	}
}
