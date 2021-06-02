package main

import (
	"strings"
	"testing"

	"github.com/airenas/tts-line/internal/pkg/mongodb"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/test"
	"github.com/stretchr/testify/assert"
)

const testAcrCfg = "acronyms:\n  url: http://server\n"
const testCliticsCfg = "clitics:\n  url: http://server\n"
const testAccenterCfg = "accenter:\n  url: http://server\n"
const testTransCfg = "transcriber:\n  url: http://server\n"
const testAMCfg = "acousticModel:\n  url: http://server\n"
const testVocCfg = "vocoder:\n  url: http://server\n"
const testCompCfg = "comparator:\n  url: http://server\n"
const testTaggerCfg = "tagger:\n  url: http://server\n"
const testValidatorCfg = "validator:\n  url: http://server\n  check:\n    min_words: 1\n"
const testConvCfg = "audioConvert:\n  url: http://server\n"

var testAllCfg = testCompCfg +
	testAccenterCfg + testTransCfg + testAMCfg + testVocCfg + testTaggerCfg + testValidatorCfg +
	testConvCfg + testAcrCfg + testCliticsCfg

var testDBSession *mongodb.SessionProvider

func initTest(t *testing.T) {
	testDBSession, _ = mongodb.NewSessionProvider("mongo://url")
}

func TestAddPartProcessors(t *testing.T) {
	partRunner := synthesizer.NewPartRunner(1)
	assert.Nil(t, addPartProcessors(partRunner, test.NewConfig(t, testAllCfg)))
}

func TestAddPartProcessors_Fail(t *testing.T) {
	partRunner := synthesizer.NewPartRunner(1)
	assert.NotNil(t, addPartProcessors(partRunner, test.NewConfig(t, trim(testAllCfg, testAcrCfg))))
	assert.NotNil(t, addPartProcessors(partRunner, test.NewConfig(t, testAcrCfg+testTransCfg+testAMCfg+testVocCfg+testCliticsCfg)))
	assert.NotNil(t, addPartProcessors(partRunner, test.NewConfig(t, testAcrCfg+testAccenterCfg+testAMCfg+testVocCfg+testCliticsCfg)))
	assert.NotNil(t, addPartProcessors(partRunner, test.NewConfig(t, testAcrCfg+testAccenterCfg+testTransCfg+testVocCfg+testCliticsCfg)))
	assert.NotNil(t, addPartProcessors(partRunner, test.NewConfig(t, testAcrCfg+testAccenterCfg+testTransCfg+testAMCfg+testCliticsCfg)))
	assert.NotNil(t, addPartProcessors(partRunner, test.NewConfig(t, testAcrCfg+testAccenterCfg+
		testTransCfg+testAMCfg+testVocCfg)))
}

func TestAddPartProcessors_NoVoc(t *testing.T) {
	partRunner := synthesizer.NewPartRunner(1)
	assert.Nil(t, addPartProcessors(partRunner, test.NewConfig(t, testAllCfg+
		"acousticModel:\n  url: http://server\n  hasVocoder: true\n")))
}

func TestAddCustomProcessors(t *testing.T) {
	initTest(t)
	syntC := &synthesizer.MainWorker{}
	assert.Nil(t, addCustomProcessors(syntC, testDBSession, test.NewConfig(t, testAllCfg)))
}

func TestAddCustomProcessors_Fail(t *testing.T) {
	initTest(t)
	syntC := &synthesizer.MainWorker{}
	assert.NotNil(t, addCustomProcessors(syntC, testDBSession, test.NewConfig(t,
		trim(testAllCfg, testCompCfg))))
	assert.NotNil(t, addCustomProcessors(syntC, testDBSession, test.NewConfig(t,
		trim(testAllCfg, testTaggerCfg))))
	assert.NotNil(t, addCustomProcessors(syntC, testDBSession, test.NewConfig(t, testCompCfg+
		trim(testAllCfg, testValidatorCfg))))
	assert.NotNil(t, addCustomProcessors(syntC, testDBSession, test.NewConfig(t,
		trim(testAllCfg, testCliticsCfg))))
	assert.NotNil(t, addCustomProcessors(syntC, testDBSession, test.NewConfig(t,
		trim(testAllCfg, testAcrCfg))))
}

func trim(all, what string) string {
	return strings.Replace(all, what, "", -1)
}
