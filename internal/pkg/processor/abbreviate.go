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
}

type abbreviator struct {
	httpWrap HTTPInvokerJSON
}

//NewAbbreviator creates new processor
func NewAbbreviator(urlStr string) (synthesizer.PartProcessor, error) {
	res := &abbreviator{}
	var err error
	res.httpWrap, err = utils.NewHTTWrap(urlStr)
	if err != nil {
		return nil, errors.Wrap(err, "Can't init http client")
	}
	return res, nil
}

func (p *abbreviator) Process(data *synthesizer.TTSDataPart) error {
	inData := mapAbbrInput(data)
	if len(inData) > 0 {
		var outData []abbrWordOutput
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

type abbrInput struct {
	Word string `json:"word,omitempty"`
	MI   string `json:"mi,omitempty"`
	ID   string `json:"id,omitempty"`
}

type abbrWordOutput struct {
	ID    string           `json:"id,omitempty"`
	Words []abbrResultWord `json:"words,omitempty"`
}

type abbrResultWord struct {
	Word      string `json:"word,omitempty"`
	WordTrans string `json:"wordTrans,omitempty"`
	Syll      string `json:"syll,omitempty"`
	UserTrans string `json:"userTrans,omitempty"`
}

func mapAbbrInput(data *synthesizer.TTSDataPart) []abbrInput {
	res := []abbrInput{}
	for i, w := range data.Words {
		tgw := w.Tagged
		if tgw.IsWord() && isAbbr(tgw.Mi, tgw.Lemma) {
			res = append(res, abbrInput{Word: tgw.Word, MI: tgw.Mi, ID: strconv.Itoa(i)})
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

func mapAbbrOutput(data *synthesizer.TTSDataPart, abbrOut []abbrWordOutput) error {
	om := make(map[int]abbrWordOutput)
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

func newWords(aw []abbrResultWord, w *synthesizer.ProcessedWord) []*synthesizer.ProcessedWord {
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
