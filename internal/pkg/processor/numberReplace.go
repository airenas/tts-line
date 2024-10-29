package processor

import (
	"fmt"
	"strings"
	"time"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/accent"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/pkg/errors"
)

// HTTPInvoker makes http call
type HTTPInvoker interface {
	InvokeText(string, interface{}) error
}

type numberReplace struct {
	httpWrap HTTPInvoker
}

// NewNumberReplace creates new processor
func NewNumberReplace(urlStr string) (synthesizer.Processor, error) {
	res := &numberReplace{}
	var err error
	res.httpWrap, err = newHTTPWrapBackoff(urlStr, time.Second*10)

	if err != nil {
		return nil, errors.Wrap(err, "can't init http client")
	}
	return res, nil
}

func (p *numberReplace) Process(data *synthesizer.TTSData) error {
	if p.skip(data) {
		goapp.Log.Info().Msg("Skip numberReplace")
		return nil
	}
	res := ""
	err := p.httpWrap.InvokeText(accent.ClearAccents(strings.Join(data.Text, " ")), &res)
	if err != nil {
		return err
	}
	data.TextWithNumbers, err = mapAccentsBack(res, data.Text)
	return err
}

func (p *numberReplace) skip(data *synthesizer.TTSData) bool {
	return data.Cfg.JustAM
}

// Info return info about processor
func (p *numberReplace) Info() string {
	return fmt.Sprintf("numberReplace(%s)", utils.RetrieveInfo(p.httpWrap))
}

type ssmlNumberReplace struct {
	httpWrap HTTPInvoker
}

// NewSSMLNumberReplace creates new processor
func NewSSMLNumberReplace(urlStr string) (synthesizer.Processor, error) {
	res := &ssmlNumberReplace{}
	var err error
	res.httpWrap, err = newHTTPWrapBackoff(urlStr, time.Second*20)

	if err != nil {
		return nil, errors.Wrap(err, "can't init http client")
	}
	return res, nil
}

func (p *ssmlNumberReplace) Process(data *synthesizer.TTSData) error {
	if p.skip(data) {
		goapp.Log.Info().Msg("Skip numberReplace")
		return nil
	}
	res := ""
	err := p.httpWrap.InvokeText(accent.ClearAccents(strings.Join(data.Text, " ")), &res)
	if err != nil {
		return err
	}
	data.TextWithNumbers, err = mapAccentsBack(res, data.Text)
	return err
}

func mapAccentsBack(new string, origArr []string) ([]string, error) {
	orig := strings.Join(origArr, " ")
	oStrs := strings.Split(orig, " ")
	nStrs := strings.Split(new, " ")
	accWrds := map[int]bool{}
	ocStrs := make([]string, len(oStrs))
	for i, s := range oStrs {
		ocStrs[i] = accent.ClearAccents(s)
		if ocStrs[i] != s {
			accWrds[i] = true
		}
	}

	alignIDs, err := utils.Align(ocStrs, nStrs)
	if err != nil {
		return nil, errors.Wrapf(err, "can't align")
	}
	for k := range accWrds {
		nID := alignIDs[k]
		if nID == -1 {
			return nil, errors.Errorf("no word alignment for %s, ID: %d", oStrs[k], k)
		}
		if nStrs[nID] != ocStrs[k] {
			return nil, errors.Errorf("no word alignment for %s, ID: %d, got %s", oStrs[k], k, nStrs[nID])
		}
		nStrs[nID] = oStrs[k]
	}
	var res []string
	wFrom, nFrom, nTo := 0, 0, 0
	for _, s := range origArr {
		l := strings.Count(s, " ") + 1
		wTo := l + wFrom
		nTo = len(nStrs)
		if wTo < len(alignIDs) {
			nTo = alignIDs[wTo]
		}
		for nTo == -1 && wTo < (len(alignIDs)-1) {
			wTo++
			nTo = alignIDs[wTo]
		}
		if nTo == -1 {
			return nil, errors.Errorf("no word alignment for %v", s)
		}
		res = append(res, strings.Join(nStrs[nFrom:nTo], " "))
		wFrom += l
		nFrom = nTo
	}
	return res, nil
}

func (p *ssmlNumberReplace) skip(data *synthesizer.TTSData) bool {
	return data.Cfg.JustAM
}

// Info return info about processor
func (p *ssmlNumberReplace) Info() string {
	return fmt.Sprintf("SSMLNumberReplace(%s)", utils.RetrieveInfo(p.httpWrap))
}
