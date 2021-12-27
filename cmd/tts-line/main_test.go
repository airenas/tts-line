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
const testObsceneCfg = "obscene:\n  url: http://server\n"

var testAllCfg = testCompCfg +
	testAccenterCfg + testTransCfg + testAMCfg + testVocCfg + testTaggerCfg + testValidatorCfg +
	testConvCfg + testAcrCfg + testCliticsCfg + testObsceneCfg

var testDBSession *mongodb.SessionProvider

func initTest(t *testing.T) {
	testDBSession, _ = mongodb.NewSessionProvider("mongo://url")
}

func TestAddPartProcessors_Custom(t *testing.T) {
	tests := []struct {
		name    string
		trimCfg string
		wantErr bool
	}{
		{name: "OK", trimCfg: "", wantErr: false},
		{name: "Obscene fail", trimCfg: testObsceneCfg, wantErr: true},
		{name: "AM fail", trimCfg: testAMCfg, wantErr: true},
		{name: "Acr fail", trimCfg: testAcrCfg, wantErr: true},
		{name: "Clitics fail", trimCfg: testCliticsCfg, wantErr: true},
		{name: "Trans fail", trimCfg: testTransCfg, wantErr: true},
		{name: "Voc fail", trimCfg: testVocCfg, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initTest(t)
			syntC := &synthesizer.MainWorker{}
			assert.Equal(t, tt.wantErr,
				addCustomProcessors(syntC, testDBSession, test.NewConfig(t, trim(testAllCfg, tt.trimCfg))) != nil)
		})
	}
}

func TestAddPartProcessors(t *testing.T) {
	tests := []struct {
		name    string
		trimCfg string
		wantErr bool
	}{
		{name: "OK", trimCfg: "", wantErr: false},
		{name: "Obscene fail", trimCfg: testObsceneCfg, wantErr: true},
		{name: "AM fail", trimCfg: testAMCfg, wantErr: true},
		{name: "Acr fail", trimCfg: testAcrCfg, wantErr: true},
		{name: "Clitics fail", trimCfg: testCliticsCfg, wantErr: true},
		{name: "Trans fail", trimCfg: testTransCfg, wantErr: true},
		{name: "Voc fail", trimCfg: testVocCfg, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initTest(t)
			partRunner := synthesizer.NewPartRunner(1)
			assert.Equal(t, tt.wantErr,
				addPartProcessors(partRunner, test.NewConfig(t, trim(testAllCfg, tt.trimCfg))) != nil)
		})
	}
}

func trim(all, what string) string {
	return strings.Replace(all, what, "", -1)
}
