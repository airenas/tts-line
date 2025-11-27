package processor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/airenas/tts-line/internal/pkg/accent"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/airenas/tts-line/internal/pkg/utils/dtw"
	"github.com/rs/zerolog/log"
)

// HTTPInvoker makes http call
type HTTPInvoker interface {
	InvokeText(context.Context, string, interface{}) error
}

type numberReplace struct {
	httpWrap HTTPInvoker
}

// NewNumberReplace creates new processor
func NewNumberReplace(urlStr string) (synthesizer.Processor, error) {
	res := &numberReplace{}
	var err error
	res.httpWrap, err = newHTTPWrapBackoff(urlStr, time.Second*20)

	if err != nil {
		return nil, fmt.Errorf("init http client: %w", err)
	}
	return res, nil
}

func (p *numberReplace) Process(ctx context.Context, data *synthesizer.TTSData) error {
	ctx, span := utils.StartSpan(ctx, "numberReplace.Process")
	defer span.End()

	if p.skip(data) {
		log.Ctx(ctx).Info().Msg("Skip numberReplace")
		return nil
	}
	res := ""
	err := p.httpWrap.InvokeText(ctx, accent.ClearAccents(strings.Join(data.Text, " ")), &res)
	if err != nil {
		return err
	}
	data.TextWithNumbers, err = mapAccentsBack(ctx, res, data.Text)
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
		return nil, fmt.Errorf("init http client: %w", err)
	}
	return res, nil
}

func (p *ssmlNumberReplace) Process(ctx context.Context, data *synthesizer.TTSData) error {
	ctx, span := utils.StartSpan(ctx, "ssmlNumberReplace.Process")
	defer span.End()

	if p.skip(data) {
		log.Ctx(ctx).Info().Msg("Skip numberReplace")
		return nil
	}
	res := ""
	err := p.httpWrap.InvokeText(ctx, accent.ClearAccents(strings.Join(data.Text, " ")), &res)
	if err != nil {
		return err
	}
	data.TextWithNumbers, err = mapAccentsBack(ctx, res, data.Text)
	return err
}

func mapAccentsBack(ctx context.Context, new string, origArr []string) ([]string, error) {
	ctx, span := utils.StartSpan(ctx, "numberReplace.mapAccentsBack")
	defer span.End()

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
	nocStrs := make([]string, len(nStrs))
	for i, s := range nStrs {
		nocStrs[i] = accent.ClearAccents(s)
	}

	alignIDs, err := dtw.Align(ctx, ocStrs, nocStrs)
	if err != nil {
		return nil, fmt.Errorf("align: %w", err)
	}

	// lId := 0
	// for i, s := range ocStrs {
	// 	var nID int
	// 	if i+1 < len(alignIDs) {
	// 		nID = alignIDs[i+1]
	// 	} else {
	// 		nID = len(nocStrs)
	// 	}
	// 	if nID == -1 {
	// 		log.Debug().Str("orig", s).Send()
	// 		continue
	// 	}
	// 	res := ""
	// 	for lId < nID {
	// 		res += nocStrs[lId] + " "
	// 		lId++
	// 	}

	// 	log.Debug().Str("orig", s).Str("new", res).Send()
	// }

	for k := range accWrds {
		nID := alignIDs[k]
		if nID == -1 {
			return nil, fmt.Errorf("no word alignment for %s, ID: %d", oStrs[k], k)
		}
		if nocStrs[nID] != ocStrs[k] {
			return nil, fmt.Errorf("no word alignment for %s, ID: %d, got %s", ocStrs[k], k, nocStrs[nID])
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
			return nil, fmt.Errorf("no word alignment for %v", s)
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
