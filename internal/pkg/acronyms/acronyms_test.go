package acronyms

import (
	"strings"
	"testing"

	"github.com/airenas/tts-line/internal/pkg/acronyms/model"
	"github.com/stretchr/testify/assert"
)

func TestFailsAcronymsOnNoData(t *testing.T) {
	_, err := New(strings.NewReader(""))
	assert.NotNil(t, err)
}

func TestFailsAcronymsOnEmptyLines(t *testing.T) {
	_, err := New(strings.NewReader("  \n   \n"))
	assert.NotNil(t, err)
}

func TestReadsAcronyms(t *testing.T) {
	prv, err := New(strings.NewReader("olia o-lia9"))
	assert.Nil(t, err)
	assert.NotNil(t, prv)
}

func TestReturnAcronyms(t *testing.T) {
	prv, err := New(strings.NewReader("ola o-lia9"))
	assert.Nil(t, err)
	res, err := prv.Process(&model.Input{Word: "ola", MI: "1"})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 1, len(res))
	if len(res) == 1 {
		assert.Equal(t, "ola", res[0].Word)
		assert.Equal(t, "olia", res[0].WordTrans)
		assert.Equal(t, "o-lia", res[0].Syll)
		assert.Equal(t, "olia9", res[0].UserTrans)
	}
}

func TestReturnAcronymsLower(t *testing.T) {
	prv, err := New(strings.NewReader("ola o-lia9"))
	assert.Nil(t, err)
	res, err := prv.Process(&model.Input{Word: "OLA", MI: "1"})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 1, len(res))
	if len(res) == 1 {
		assert.Equal(t, "OLA", res[0].Word)
		assert.Equal(t, "olia", res[0].WordTrans)
		assert.Equal(t, "o-lia", res[0].Syll)
		assert.Equal(t, "olia9", res[0].UserTrans)
	}
}

func TestReturnAcronymsLower2(t *testing.T) {
	prv, err := New(strings.NewReader("OLA o-lia9"))
	assert.Nil(t, err)
	res, err := prv.Process(&model.Input{Word: "ola", MI: "1"})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 1, len(res))
	if len(res) == 1 {
		assert.Equal(t, "ola", res[0].Word)
		assert.Equal(t, "olia", res[0].WordTrans)
		assert.Equal(t, "o-lia", res[0].Syll)
		assert.Equal(t, "olia9", res[0].UserTrans)
	}
}

func TestReturnAcronymsMultiple(t *testing.T) {
	prv, err := New(strings.NewReader("ola o3 li-a9"))
	assert.Nil(t, err)
	res, err := prv.Process(&model.Input{Word: "ola", MI: "1"})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 2, len(res))
	if len(res) == 2 {
		assert.Equal(t, "o", res[0].Word)
		assert.Equal(t, "o", res[0].WordTrans)
		assert.Equal(t, "o", res[0].Syll)
		assert.Equal(t, "o3", res[0].UserTrans)

		assert.Equal(t, "lia", res[1].Word)
		assert.Equal(t, "lia", res[1].WordTrans)
		assert.Equal(t, "li-a", res[1].Syll)
		assert.Equal(t, "lia9", res[1].UserTrans)
	}
}
