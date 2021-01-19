package main

import (
	"testing"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/test"
	"github.com/stretchr/testify/assert"
)

const testAbbrCfg = "abbreviator:\n  url: http://server\n"
const testAccenterCfg = "accenter:\n  url: http://server\n"
const testTransCfg = "transcriber:\n  url: http://server\n"
const testAMCfg = "acousticModel:\n  url: http://server\n"
const testVocCfg = "vocoder:\n  url: http://server\n"

func TestAddPartProcessors(t *testing.T) {
	partRunner := synthesizer.NewPartRunner(1)
	assert.Nil(t, addPartProcessors(partRunner, test.NewConfig(testAbbrCfg+testAccenterCfg+testTransCfg+testAMCfg+testVocCfg)))
}

func TestAddPartProcessors_Fail(t *testing.T) {
	partRunner := synthesizer.NewPartRunner(1)
	assert.NotNil(t, addPartProcessors(partRunner, test.NewConfig(testAccenterCfg+testTransCfg+testAMCfg+testVocCfg)))
	assert.NotNil(t, addPartProcessors(partRunner, test.NewConfig(testAbbrCfg+testTransCfg+testAMCfg+testVocCfg)))
	assert.NotNil(t, addPartProcessors(partRunner, test.NewConfig(testAbbrCfg+testAccenterCfg+testAMCfg+testVocCfg)))
	assert.NotNil(t, addPartProcessors(partRunner, test.NewConfig(testAbbrCfg+testAccenterCfg+testTransCfg+testVocCfg)))
	assert.NotNil(t, addPartProcessors(partRunner, test.NewConfig(testAbbrCfg+testAccenterCfg+testTransCfg+testAMCfg)))
}

func TestAddPartProcessors_NoVoc(t *testing.T) {
	partRunner := synthesizer.NewPartRunner(1)
	assert.Nil(t, addPartProcessors(partRunner, test.NewConfig(testAbbrCfg+testAccenterCfg+testTransCfg+
		"acousticModel:\n  url: http://server\n  hasVocoder: true\n")))
}
