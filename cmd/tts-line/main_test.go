package main

import (
	"strings"
	"testing"

	"github.com/airenas/tts-line/internal/pkg/mongodb"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAcrCfg = "acronyms:\n  url: http://server\n"
const testCliticsCfg = "clitics:\n  url: http://server\n"
const testAccenterCfg = "accenter:\n  url: http://server\n"
const testTransCfg = "transcriber:\n  url: http://server\n"
const testAMCfg = "acousticModel:\n  url: http://server\n"
const testVocCfg = "vocoder:\n  url: http://server\n"
const testCompCfg = "comparator:\n  url: http://server\n"
const testTaggerCfg = "tagger:\n  url: http://server\n"
const testWordTaggerCfg = "wordTagger:\n  url: http://server\n"
const testURLReaderCfg = "urlReader:\n  url: http://server\n"
const testValidatorCfg = "validator:\n  maxChars: 100\n"
const testConvCfg = "audioConvert:\n  url: http://server\n"
const testObsceneCfg = "obscene:\n  url: http://server\n"
const testCleanCfg = "clean:\n  url: http://cl.su\n"
const testTransliteratorCfg = "transliterator:\n  url: http://transl.su\n"
const testNormalizeCfg = "normalize:\n  url: http://norm.su\n"
const testNumberReplaceCfg = "numberReplace:\n  url: http://nr.su\n"
const testSuffixLoaderCfg = "suffixLoader:\n  path: ./\n"

var testAllCfg = testCompCfg +
	testAccenterCfg + testTransCfg + testAMCfg + testVocCfg + testTaggerCfg + testWordTaggerCfg +
	testValidatorCfg +
	testConvCfg + testAcrCfg + testCliticsCfg + testObsceneCfg + testCleanCfg +
	testNumberReplaceCfg + testSuffixLoaderCfg + testNormalizeCfg + testTransliteratorCfg + testURLReaderCfg

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
		{name: "Validator fail", trimCfg: testValidatorCfg, wantErr: true},
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

func TestAddSSMLProcessors(t *testing.T) {
	mw := synthesizer.MainWorker{}
	err := addSSMLProcessors(&mw, &mongodb.SessionProvider{}, test.NewConfig(t, testAllCfg))
	assert.Nil(t, err)
	info := mw.GetSSMLProcessorsInfo()
	req := []string{"addMetrics",
		"SSMLValidator(100)",
		"saver(originalSSML)",
		"SSMLPartRunner",
		"cleaner(HTTPBackoff(HTTPWrap(http://cl.su, tm: 10s)))",
		"normalizer(HTTPBackoff(HTTPWrap(http://norm.su, tm: 10s)))",
		"numberReplace(HTTPBackoff(HTTPWrap(http://nr.su, tm: 20s)))",
		"SSMLTagger(", "joinSSMLAudio(audioLoader(./))",
		"audioConverter", "addMetrics"}
	infos := strings.Split(info, "\n")
	pos := 0
	for _, rs := range req {
		was := false
		for ; pos < len(infos); pos++ {
			if strings.Contains(infos[pos], rs) {
				was = true
				break
			}
		}
		require.True(t, was, "no `%s` in [%s]", rs, strings.Join(infos, ">"))
	}
}

func TestAddProcessors(t *testing.T) {
	mw := synthesizer.MainWorker{}
	err := addProcessors(&mw, &mongodb.SessionProvider{}, test.NewConfig(t, testAllCfg))
	assert.Nil(t, err)
	info := mw.GetProcessorsInfo()
	req := []string{"addMetrics",
		"saver(original)",
		"cleaner(HTTPBackoff(HTTPWrap(http://cl.su, tm: 10s)))",
		"normalizer(HTTPBackoff(HTTPWrap(http://norm.su, tm: 10s)))",
		"saver(cleaned)",
		"numberReplace(HTTPBackoff(HTTPWrap(http://nr.su, tm: 20s)))",
		"tagger(",
		"urlReplacer(",
		"transliterator(",
		"saver(normalized)",
		"joinAudio(audioLoader(./))",
		"audioConverter", "addMetrics"}
	infos := strings.Split(info, "\n")
	pos := 0
	for _, rs := range req {
		was := false
		for ; pos < len(infos); pos++ {
			if strings.Contains(infos[pos], rs) {
				was = true
				break
			}
		}
		require.True(t, was, "no `%s` in [%s]", rs, strings.Join(infos, ">"))
	}
}

func trim(all, what string) string {
	return strings.Replace(all, what, "", -1)
}
