package processor

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/rs/zerolog/log"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/airenas/tts-line/pkg/ssml"
)

type normalizer struct {
	httpWrap HTTPInvokerJSON
}

// NewNormalizer creates new text normalize processor
func NewNormalizer(urlStr string) (synthesizer.Processor, error) {
	res := &normalizer{}
	var err error
	res.httpWrap, err = newHTTPWrapBackoff(urlStr, time.Second*10)
	if err != nil {
		return nil, fmt.Errorf("can't init http client: %w", err)
	}
	return res, nil
}

func (p *normalizer) Process(ctx context.Context, data *synthesizer.TTSData) error {
	if p.skip(data) {
		log.Ctx(ctx).Info().Msg("Skip normalize")
		return nil
	}
	defer goapp.Estimate("Normalize")()

	if len(data.CleanedText) != len(data.OriginalTextParts) {
		return fmt.Errorf("mismatch in parts len %d vs %d", len(data.CleanedText), len(data.OriginalTextParts))
	}

	inp, err := mapNormalizeInput(data)
	if err != nil {
		return err
	}
	// txt := strings.Join(data.CleanedText, " ")
	// utils.LogData(ctx, "Input", txt, nil)
	inData := []*normRequestData{&normRequestData{Orig: inp}}
	var output []*normResponseData
	err = p.httpWrap.InvokeJSON(ctx, inData, &output)
	if err != nil {
		if len(output) == 0 {
			return fmt.Errorf("normalize: %w", err)
		}
		return fmt.Errorf("normalize (%s): %w", output[0].Err, err)
	}

	if len(output) != 1 {
		return fmt.Errorf("normalize: unexpected output len %d", len(output))
	}

	data.NormalizedText, err = processNormalizedOutput(output[0], data)
	if err != nil {
		return err
	}
	utils.LogData(ctx, "Output", strings.Join(data.NormalizedText, " "), nil)
	return nil
}

func mapNormalizeInput(data *synthesizer.TTSData) ([]*normText, error) {
	res := []*normText{}
	for i := range data.CleanedText {
		if i >= len(data.OriginalTextParts) {
			return nil, fmt.Errorf("mismatch in parts len %d vs %d", len(data.CleanedText), len(data.OriginalTextParts))
		}
		nt := &normText{
			Text: getPartText(data, i),
			Type: mapNormalizeType(data.OriginalTextParts[i]),
		}
		res = append(res, nt)
	}
	return res, nil
}

func getPartText(data *synthesizer.TTSData, i int) string {
	if isFixedText(data.OriginalTextParts[i]) {
		return data.OriginalTextParts[i].Text + " "
	}
	return data.CleanedText[i] + " "
}

func mapNormalizeType(part *synthesizer.TTSTextPart) normTextType {
	if isFixedText(part) {
		return normTextTypeFixed
	}
	return normTextTypePlain
}

func isFixedText(part *synthesizer.TTSTextPart) bool {
	if part.InterpretAs != ssml.InterpretAsTypeUnset {
		return true
	}
	if part.Accented != "" || part.UserOEPal != "" {
		return true
	}
	return false
}

func processNormalizedOutput(output *normResponseData, data *synthesizer.TTSData) ([]string, error) {
	input := data.CleanedText

	if len(output.Rep) == 0 {
		return input, nil
	}
	resR := make([][]rune, len(input))
	fixed := make([]bool, len(input))
	for i := range input {
		resR[i] = []rune(getPartText(data, i))
		fixed[i] = isFixedText(data.OriginalTextParts[i])
	}

	changes := output.Rep
	sort.SliceStable(changes, func(i, j int) bool {
		if changes[i].Priority == changes[j].Priority {
			return changes[i].Beg < changes[j].Beg
		}
		return changes[i].Priority < changes[j].Priority
	})

	shift := 0
	lastPriority := -1
	for _, rep := range changes {
		if lastPriority != rep.Priority {
			shift = 0
			lastPriority = rep.Priority
		}
		atI, fromI, err := getNextStr(resR, rep.Beg-shift)
		if err != nil {
			return nil, fmt.Errorf("err replace %v: %w", rep, err)
		}
		if fixed[fromI] {
			continue
		}
		rns := resR[fromI]
		repRns := []rune(rep.Text)
		rnsNew := append([]rune{}, rns[:atI]...)
		rnsNew = append(rnsNew, repRns...)
		end := (rep.End - rep.Beg) + atI
		if end < len(rns) {
			rnsNew = append(rnsNew, rns[end:]...)
		}
		resR[fromI] = rnsNew
		rem := end - len(rns)
		for rem > 0 {
			fromI++
			if len(resR[fromI]) > rem {
				resR[fromI] = resR[fromI][rem:]
				rem = 0
			} else {
				rem -= len(resR[fromI])
				resR[fromI] = nil
			}
		}
		shift += (rep.End - rep.Beg) - len(repRns)
	}

	res := make([]string, len(resR))
	for i := range input {
		if fixed[i] {
			res[i] = input[i]
			continue
		}
		r := resR[i]
		if len(r) > 0 && r[len(r)-1] == ' ' {
			r = r[:len(r)-1]
		}
		res[i] = string(r)
	}
	return res, nil
}

func getNextStr(resR [][]rune, at int) (int, int, error) {
	for i := 0; i < len(resR); i++ {
		res := resR[i]
		l := len(res)
		// if i < len(resR)-1 {
		// 	l++ // compensate strings.Join(data.CleanedText, " ")
		// }
		if at >= l {
			at -= l
		} else {
			return at, i, nil
		}
	}
	return 0, 0, fmt.Errorf("wrong start pos at %d", at)
}

type normRequestData struct {
	Orig []*normText `json:"org"`
}

type normTextType string

const (
	normTextTypePlain normTextType = "plain"
	normTextTypeFixed normTextType = "fixed"
)

type normText struct {
	Text string       `json:"text,omitempty"`
	Type normTextType `json:"type,omitempty"`
}

type normResponseData struct {
	Err string                 `json:"err"`
	Org []*normText            `json:"org"`
	Rep []*normResponseDataRep `json:"rep"`
	Res string                 `json:"res"`
}

type normResponseDataRep struct {
	Beg      int    `json:"beg"`
	End      int    `json:"end"`
	Priority int    `json:"priority"`
	Text     string `json:"text"`
}

func (p *normalizer) skip(data *synthesizer.TTSData) bool {
	return data.Cfg.JustAM
}

// Info return info about processor
func (p *normalizer) Info() string {
	return fmt.Sprintf("normalizer(%s)", utils.RetrieveInfo(p.httpWrap))
}
