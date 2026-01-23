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
	httpWrap  HTTPInvoker
	urlFinder *urlFinder
}

type replaceData struct {
	data []*textURLReplace
}

type textURLReplace struct {
	orig        string
	links       []string
	placeholder string
	replaced    string
}

func (t *replaceData) textArray() []string {
	var parts []string
	for _, d := range t.data {
		parts = append(parts, d.replaced)
	}
	return parts
}

func (t *replaceData) text() string {
	return strings.Join(t.textArray(), " ")
}

func (t *replaceData) restore(textWithNumbers []string) ([]string, error) {
	var res []string
	if len(textWithNumbers) != len(t.data) {
		return nil, fmt.Errorf("length mismatch: got %d, expected %d", len(textWithNumbers), len(t.data))
	}
	for i, part := range textWithNumbers {
		r := t.data[i]
		final := part
		for _, link := range r.links {
			final = strings.Replace(final, r.placeholder, link, 1)
		}
		res = append(res, final)
	}
	return res, nil
}

// NewNumberReplace creates new processor
func NewNumberReplace(urlStr string) (synthesizer.Processor, error) {
	finder, err := NewURLFinder()
	if err != nil {
		return nil, fmt.Errorf("init url finder: %w", err)
	}
	res := &numberReplace{urlFinder: finder}
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

	noURLS, err := removeURLs(ctx, p.urlFinder, data.NormalizedText)
	if err != nil {
		return fmt.Errorf("remove URLs: %w", err)
	}
	res := ""
	err = p.httpWrap.InvokeText(ctx, accent.ClearAccents(noURLS.text()), &res)
	if err != nil {
		return err
	}
	textWithNumbers, err := mapAccentsBack(ctx, res, noURLS.textArray())
	if err != nil {
		return fmt.Errorf("map accents back: %w", err)
	}
	data.TextWithNumbers, err = noURLS.restore(textWithNumbers)
	if err != nil {
		return fmt.Errorf("restore URLs: %w", err)
	}

	return err
}

func removeURLs(ctx context.Context, replacer *urlFinder, s []string) (*replaceData, error) {
	res := &replaceData{}
	res.data = make([]*textURLReplace, 0, len(s))
	for _, part := range s {
		r, err := removeURL(ctx, replacer, part)
		if err != nil {
			return nil, err
		}
		res.data = append(res.data, r)
	}
	return res, nil
}

func removeURL(_ctx context.Context, replacer *urlFinder, part string) (*textURLReplace, error) {
	placeholder := "URL"
	for strings.Contains(part, placeholder) {
		placeholder += "_"
	}
	res := &textURLReplace{orig: part, placeholder: placeholder}
	res.links, res.replaced = replacer.replaceAll(part, placeholder)
	return res, nil
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
	err := p.httpWrap.InvokeText(ctx, accent.ClearAccents(strings.Join(data.CleanedText, " ")), &res)
	if err != nil {
		return err
	}
	data.TextWithNumbers, err = mapAccentsBack(ctx, res, data.CleanedText)
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
