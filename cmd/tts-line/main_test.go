package main

import (
	"testing"

	"github.com/airenas/tts-line/internal/pkg/mongodb"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/test"
	"github.com/stretchr/testify/assert"
)

const testAcrCfg = "acronyms:\n  url: http://server\n"
const testAccenterCfg = "accenter:\n  url: http://server\n"
const testTransCfg = "transcriber:\n  url: http://server\n"
const testAMCfg = "acousticModel:\n  url: http://server\n"
const testVocCfg = "vocoder:\n  url: http://server\n"
const testCompCfg = "comparator:\n  url: http://server\n"
const testTaggerCfg = "tagger:\n  url: http://server\n"
const testValidatorCfg = "validator:\n  url: http://server\n  check:\n    min_words: 1\n"
const testConvCfg = "audioConvert:\n  url: http://server\n"

var testDBSession *mongodb.SessionProvider

func initTest(t *testing.T) {
	testDBSession, _ = mongodb.NewSessionProvider("mongo://url")
}

func TestAddPartProcessors(t *testing.T) {
	partRunner := synthesizer.NewPartRunner(1)
	assert.Nil(t, addPartProcessors(partRunner, test.NewConfig(t, testAcrCfg+testAccenterCfg+
		testTransCfg+testAMCfg+testVocCfg)))
}

func TestAddPartProcessors_Fail(t *testing.T) {
	partRunner := synthesizer.NewPartRunner(1)
	assert.NotNil(t, addPartProcessors(partRunner, test.NewConfig(t, testAccenterCfg+testTransCfg+testAMCfg+testVocCfg)))
	assert.NotNil(t, addPartProcessors(partRunner, test.NewConfig(t, testAcrCfg+testTransCfg+testAMCfg+testVocCfg)))
	assert.NotNil(t, addPartProcessors(partRunner, test.NewConfig(t, testAcrCfg+testAccenterCfg+testAMCfg+testVocCfg)))
	assert.NotNil(t, addPartProcessors(partRunner, test.NewConfig(t, testAcrCfg+testAccenterCfg+testTransCfg+testVocCfg)))
	assert.NotNil(t, addPartProcessors(partRunner, test.NewConfig(t, testAcrCfg+testAccenterCfg+testTransCfg+testAMCfg)))
}

func TestAddPartProcessors_NoVoc(t *testing.T) {
	partRunner := synthesizer.NewPartRunner(1)
	assert.Nil(t, addPartProcessors(partRunner, test.NewConfig(t, testAcrCfg+testAccenterCfg+testTransCfg+
		"acousticModel:\n  url: http://server\n  hasVocoder: true\n")))
}

func TestAddCustomProcessors(t *testing.T) {
	initTest(t)
	syntC := &synthesizer.MainWorker{}
	assert.Nil(t, addCustomProcessors(syntC, testDBSession, test.NewConfig(t, testCompCfg+
		testAccenterCfg+testTransCfg+testAMCfg+testVocCfg+testTaggerCfg+testValidatorCfg+
		testConvCfg+testAcrCfg)))
}

func TestAddCustomProcessors_Fail(t *testing.T) {
	initTest(t)
	syntC := &synthesizer.MainWorker{}
	assert.NotNil(t, addCustomProcessors(syntC, testDBSession, test.NewConfig(t,
		testAccenterCfg+testTransCfg+testAMCfg+testVocCfg+testTaggerCfg+testValidatorCfg+
			testConvCfg+testAcrCfg)))
	assert.NotNil(t, addCustomProcessors(syntC, testDBSession, test.NewConfig(t, testCompCfg+
		testAccenterCfg+testTransCfg+testAMCfg+testVocCfg+testValidatorCfg+
		testConvCfg+testAcrCfg)))
	assert.NotNil(t, addCustomProcessors(syntC, testDBSession, test.NewConfig(t, testCompCfg+
		testAccenterCfg+testTransCfg+testAMCfg+testVocCfg+testTaggerCfg+
		testConvCfg+testAcrCfg)))
	assert.NotNil(t, addCustomProcessors(syntC, testDBSession, test.NewConfig(t, testCompCfg+
		testAccenterCfg+testTransCfg+testAMCfg+testVocCfg+testTaggerCfg+testValidatorCfg+
		testAcrCfg)))
	assert.NotNil(t, addCustomProcessors(syntC, nil, test.NewConfig(t, testCompCfg+
		testAccenterCfg+testTransCfg+testAMCfg+testVocCfg+testTaggerCfg+testValidatorCfg+
		testConvCfg+testAcrCfg)))
	assert.NotNil(t, addCustomProcessors(syntC, testDBSession, test.NewConfig(t, testCompCfg+
		testAccenterCfg+testTransCfg+testAMCfg+testVocCfg+testTaggerCfg+testValidatorCfg+
		testConvCfg)))
}
