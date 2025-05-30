package processor

import (
	"context"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// HTTPInvokerJSON invoker for json input
type HTTPInvokerJSON interface {
	InvokeJSON(context.Context, interface{}, interface{}) error
	InvokeJSONU(ctx context.Context, URL string, dataIn interface{}, dataOut interface{}) error
	InvokeText(context.Context, string, interface{}) error
}

type acronyms struct {
	httpWrap HTTPInvokerJSON
}

// NewAcronyms creates new processor
func NewAcronyms(urlStr string) (synthesizer.PartProcessor, error) {
	res := &acronyms{}
	var err error
	res.httpWrap, err = newHTTPWrapBackoff(urlStr, time.Second*10)
	if err != nil {
		return nil, errors.Wrap(err, "can't init http client")
	}
	return res, nil
}

func (p *acronyms) Process(ctx context.Context, data *synthesizer.TTSDataPart) error {
	ctx, span := utils.StartSpan(ctx, "acronyms.Process")
	defer span.End()

	if p.skip(data) {
		log.Ctx(ctx).Info().Msg("Skip acronyms")
		return nil
	}
	inData := mapAbbrInput(data)
	if len(inData) > 0 {
		var outData []acrWordOutput
		err := p.httpWrap.InvokeJSON(ctx, inData, &outData)
		if err != nil {
			return err
		}
		return mapAbbrOutput(data, outData)
	} else {
		log.Ctx(ctx).Debug().Msg("Skip abbreviation - no data in")
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
		if tgw.IsWord() &&
			w.Obscene || (!isAccented(w) && // do not do acronyms change if user has provided accent
			!hasUserTranscriptions(w) &&
			isAbbr(tgw.Mi, tgw.Lemma)) {
			res = append(res, acrInput{Word: tgw.Word, MI: tgw.Mi, ID: strconv.Itoa(i)})
		}
	}
	return res
}

func hasUserTranscriptions(w *synthesizer.ProcessedWord) bool {
	return w.UserTranscription != "" || w.UserSyllables != "" || w.UserAccent != 0
}

func isAccented(w *synthesizer.ProcessedWord) bool {
	return w.UserAccent != 0 || (w.TextPart != nil && w.TextPart.Accented != "")
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
