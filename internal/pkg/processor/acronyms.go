package processor

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/pkg/errors"
)

//HTTPInvokerJSON invoker for json input
type HTTPInvokerJSON interface {
	InvokeJSON(interface{}, interface{}) error
	InvokeJSONU(URL string, dataIn interface{}, dataOut interface{}) error
}

type acronyms struct {
	httpWrap HTTPInvokerJSON
}

//NewAcronyms creates new processor
func NewAcronyms(urlStr string) (synthesizer.PartProcessor, error) {
	res := &acronyms{}
	var err error
	res.httpWrap, err = utils.NewHTTWrap(urlStr)
	if err != nil {
		return nil, errors.Wrap(err, "Can't init http client")
	}
	return res, nil
}

func (p *acronyms) Process(data *synthesizer.TTSDataPart) error {
	if p.skip(data) {
		goapp.Log.Info("Skip acronyms")
		return nil
	}
	inData := mapAbbrInput(data)
	if len(inData) > 0 {
		var outData []acrWordOutput
		err := p.httpWrap.InvokeJSON(inData, &outData)
		if err != nil {
			return err
		}
		mapAbbrOutput(data, outData)
	} else {
		goapp.Log.Debug("Skip abbreviation - no data in")
	}
	return nil
}

type acrInput struct {
	Word string `json:"word,omitempty"`
	MI   string `json:"mi,omitempty"`
	ID   string `json:"id,omitempty"`
}

type acrWordOutput struct {
	ID    string          `json:"id,omitempty"`
	Words []acrResultWord `json:"words,omitempty"`
}

type acrResultWord struct {
	Word      string `json:"word,omitempty"`
	WordTrans string `json:"wordTrans,omitempty"`
	Syll      string `json:"syll,omitempty"`
	UserTrans string `json:"userTrans,omitempty"`
}

func mapAbbrInput(data *synthesizer.TTSDataPart) []acrInput {
	res := []acrInput{}
	for i, w := range data.Words {
		tgw := w.Tagged
		if tgw.IsWord() && isAbbr(tgw.Mi, tgw.Lemma) {
			res = append(res, acrInput{Word: tgw.Word, MI: tgw.Mi, ID: strconv.Itoa(i)})
		}
	}
	return res
}

func isAbbr(mi, lemma string) bool {
	return mi != "" && (strings.HasPrefix(mi, "X") || strings.HasPrefix(mi, "Y") || mayBeAbbr(mi, lemma))
}

func mayBeAbbr(mi, lemma string) bool {
	return len(lemma) >= 2 && strings.HasPrefix(mi, "N") && allUpper(lemma)
}

func allUpper(lemma string) bool {
	for _, r := range lemma {
		if unicode.IsLetter(r) && unicode.IsLower(r) {
			return false
		}
	}
	return len(lemma) > 0
}

func mapAbbrOutput(data *synthesizer.TTSDataPart, abbrOut []acrWordOutput) error {
	om := make(map[int]acrWordOutput)
	for _, abbr := range abbrOut {
		iID, err := strconv.Atoi(abbr.ID)
		if err != nil {
			return errors.Wrapf(err, "Can't parse ID %s", abbr.ID)
		}
		om[iID] = abbr
	}
	res := make([]*synthesizer.ProcessedWord, 0)
	for i, w := range data.Words {
		abbr, ok := om[i]
		if ok {
			res = append(res, newWords(abbr.Words, w)...)
		} else {
			res = append(res, w)
		}
	}
	data.Words = res
	return nil
}

func newWords(aw []acrResultWord, w *synthesizer.ProcessedWord) []*synthesizer.ProcessedWord {
	res := []*synthesizer.ProcessedWord{}
	for i, r := range aw {
		wd := synthesizer.ProcessedWord{}
		wd.Tagged.Word = r.Word
		wd.Tagged.Mi = w.Tagged.Mi
		if i == 0 {
			wd.Tagged.Lemma = w.Tagged.Lemma
		}
		wd.UserTranscription = r.UserTrans
		wd.UserSyllables = r.Syll
		wd.TranscriptionWord = r.WordTrans
		res = append(res, &wd)
	}
	return res
}

func (p *acronyms) skip(data *synthesizer.TTSDataPart) bool {
	return data.Cfg.JustAM
}
