package processor

import (
	"os"
	"strings"
	"testing"

	"github.com/petergtz/pegomock"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/stretchr/testify/assert"

	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
)

func TestNewValidator(t *testing.T) {
	initTestJSON(t)
	pr, err := NewValidator(newTestConfig("url: http://server\ncheck:\n  min_words: 1"))
	assert.Nil(t, err)
	assert.NotNil(t, pr)
}

func TestNewValidator_Fails(t *testing.T) {
	initTestJSON(t)
	pr, err := NewValidator(nil)
	assert.NotNil(t, err)
	assert.Nil(t, pr)

	_, err = NewValidator(newTestConfig(""))
	assert.NotNil(t, err)

	_, err = NewValidator(newTestConfig("url: "))
	assert.NotNil(t, err)

	_, err = NewValidator(newTestConfig("url: http://server\n"))
	assert.NotNil(t, err)
}

func TestInitChecks(t *testing.T) {
	ch, err := initChecks(newTestConfig("min_words: 1\nmax_words: 10\nno_numbers: 1\nprofanity: 1"))
	assert.Nil(t, err)
	assert.Equal(t, 4, len(ch))
	assert.Equal(t, 1, ch[0].Value)
	assert.Equal(t, 10, ch[1].Value)
	assert.Equal(t, "no_numbers", ch[2].ID)
	assert.Equal(t, "profanity", ch[3].ID)
}

func TestInitChecks_Ignore(t *testing.T) {
	ch, err := initChecks(newTestConfig("min_words: 0"))
	assert.Nil(t, err)
	assert.Equal(t, 0, len(ch))
}

func TestInvokeValidator(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewValidator(newTestConfig("url: http://server\ncheck:\n  min_words: 1"))
	assert.NotNil(t, pr)
	pr.(*validator).httpWrap = httpJSONMock
	d := synthesizer.TTSData{}
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "word"}})
	pegomock.When(httpJSONMock.InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())).Then(
		func(params []pegomock.Param) pegomock.ReturnValues {
			assert.Equal(t, "word", params[0].(*valInput).Words.List[0].Word)
			assert.Equal(t, "min_words", params[0].(*valInput).Checks[0].ID)
			*params[1].(*valOutput) = valOutput{List: []api.ValidateFailure{api.ValidateFailure{Check: api.Check{ID: "olia"}, FailingPosition: 1}}}
			return []pegomock.ReturnValue{nil}
		})
	err := pr.Process(&d)
	assert.Nil(t, err)
	assert.Equal(t, "olia", d.ValidationFailures[0].Check.ID)
}

func TestInvokeValidator_Fail(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewValidator(newTestConfig("url: http://server\ncheck:\n  min_words: 1"))
	assert.NotNil(t, pr)
	pr.(*validator).httpWrap = httpJSONMock
	d := synthesizer.TTSData{}
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "word"}})
	pegomock.When(httpJSONMock.InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())).ThenReturn(errors.New("haha"))
	err := pr.Process(&d)
	assert.NotNil(t, err)
}

func TestMapValInput(t *testing.T) {
	pr, _ := NewValidator(newTestConfig("url: http://server\ncheck:\n  min_words: 1"))
	assert.NotNil(t, pr)
	prv := pr.(*validator)

	d := synthesizer.TTSData{}
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: "!"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "v2"}})
	inp := prv.mapValidatorInput(&d)
	assert.Equal(t, "v1", inp.Words.List[0].Word)
	assert.Equal(t, "v2", inp.Words.List[1].Word)
	assert.Equal(t, prv.checks, inp.Checks)
}

func TestMapValInput_FromConfig(t *testing.T) {
	os.Setenv("VALIDATOR_CHECK_MAX_WORDS", "300")
	pr, _ := NewValidator(newTestConfig("url: http://server\ncheck:\n  max_words: 10"))
	assert.NotNil(t, pr)
	prv := pr.(*validator)

	assert.Equal(t, 1, len(prv.checks))
	assert.Equal(t, 300, prv.checks[0].Value)
}

func newTestConfig(yaml string) *viper.Viper {
	res := viper.New()
	res.SetConfigType("yaml")
	res.ReadConfig(strings.NewReader(yaml))
	return res
}
