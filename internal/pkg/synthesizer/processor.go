package synthesizer

import (
	"fmt"
	"strings"
	"time"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/airenas/tts-line/internal/pkg/accent"
	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/airenas/tts-line/pkg/ssml"
)

// Processor interface
type Processor interface {
	Process(*TTSData) error
}

// MainWorker does synthesis work
type MainWorker struct {
	processors      []Processor
	ssmlProcessors  []Processor
	AllowCustomCode bool
}

// Work is main method
func (mw *MainWorker) Work(input *api.TTSRequestConfig) (*api.Result, error) {
	data := &TTSData{}
	data.OriginalText = input.Text
	data.Input = input
	data.Cfg.Input = input
	data.Cfg.Type = SSMLNone
	data.Cfg.Speed = input.Speed
	data.Cfg.Voice = input.Voice
	data.RequestID = input.RequestID
	data.AudioSuffix = input.AudioSuffix
	if input.RequestID == "" {
		data.RequestID = uuid.NewString()
	}
	if mw.AllowCustomCode {
		tryCustomCode(data)
	}
	if len(input.SSMLParts) > 0 {
		data.Cfg.Type = SSMLMain
		var err error
		data.SSMLParts, err = makeSSMLParts(input)
		if err != nil {
			return nil, err
		}
		if err := mw.processAll(mw.ssmlProcessors, data); err != nil {
			return nil, err
		}
	} else {
		if err := mw.processAll(mw.processors, data); err != nil {
			return nil, err
		}
	}
	return mapResult(data)
}

func makeSSMLParts(input *api.TTSRequestConfig) ([]*TTSData, error) {
	var res []*TTSData
	for _, p := range input.SSMLParts {
		switch pc := p.(type) {
		case *ssml.Text:
			data := &TTSData{}
			data.OriginalTextParts = makeTextParts(pc.Texts)
			data.Input = input
			data.Cfg.Input = input
			data.Cfg.Speed = pc.Speed
			data.Cfg.Voice = pc.Voice
			data.Cfg.Type = SSMLText
			data.RequestID = input.RequestID
			if input.RequestID == "" {
				data.RequestID = uuid.NewString()
			}
			res = append(res, data)
		case *ssml.Pause:
			data := &TTSData{}
			data.Cfg.PauseDuration = pc.Duration
			data.Cfg.Type = SSMLPause
			res = append(res, data)
		default:
			return nil, errors.Errorf("unknown SSML part type %T", pc)
		}
	}
	return res, nil
}

func makeTextParts(textPart []ssml.TextPart) []*TTSTextPart {
	res := []*TTSTextPart{}
	for _, tp := range textPart {
		res = append(res, &TTSTextPart{Text: tp.Text, Accented: tp.Accented, Syllables: tp.Syllables, UserOEPal: tp.UserOEPal})
	}
	return res
}

// Add adds a processor to the end
func (mw *MainWorker) Add(pr Processor) {
	mw.processors = append(mw.processors, pr)
}

// AddSSML adds a SSML processor to the end
func (mw *MainWorker) AddSSML(pr Processor) {
	mw.ssmlProcessors = append(mw.ssmlProcessors, pr)
}

func (mw *MainWorker) processAll(processors []Processor, data *TTSData) error {
	for _, pr := range processors {
		err := pr.Process(data)
		if err != nil {
			return err
		}
	}
	return nil
}

func mapResult(data *TTSData) (*api.Result, error) {
	res := &api.Result{}
	res.AudioAsString = data.AudioMP3
	if data.Input.OutputTextFormat != api.TextNone {
		if data.Input.AllowCollectData {
			res.RequestID = data.RequestID
		}
		if data.Input.OutputTextFormat == api.TextNormalized {
			res.Text = strings.Join(data.TextWithNumbers, " ")
		} else if data.Input.OutputTextFormat == api.TextTranscribed {
			res.Text = mapTranscribed(data)
		} else if data.Input.OutputTextFormat == api.TextAccented {
			var err error
			res.Text, err = mapAccentedText(data)
			if err != nil {
				return nil, err
			}
		} else if data.Input.OutputTextFormat == api.TextNone {
		} else {
			return nil, errors.Errorf("can't process OutputTextFormat %s", data.Input.OutputTextFormat.String())
		}
	}
	if data.Input.SpeechMarkTypes[api.SpeechMarkTypeWord] {
		var err error
		res.SpeechMarks, err = mapSpeechMarks(data)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

type wordMapData struct {
	pw    *ProcessedWord
	part  *TTSDataPart
	start time.Duration
	shift int
}

func mapSpeechMarks(data *TTSData) ([]*api.SpeechMark, error) {
	if len(data.SSMLParts) == 0 {
		res, _, err := mapSpeechMarksInt(data, 0)
		return res, err
	}

	var res []*api.SpeechMark
	from := time.Duration(0)
	for _, p := range data.SSMLParts {
		pRes, dur, err := mapSpeechMarksInt(p, from)
		if err != nil {
			return nil, err
		}
		res = append(res, pRes...)
		from += dur
	}
	return res, nil
}

func mapSpeechMarksInt(data *TTSData, from time.Duration) ([]*api.SpeechMark, time.Duration, error) {
	text := strings.Join(data.Text, " ")
	if len(text) == 0 {
		return nil, 0, nil
	}
	originalWords := strings.Fields(accent.ClearAccents(text))
	words, maps := collectWords(data.Parts)
	aligned, err := utils.Align(originalWords, words)
	if err != nil {
		return nil, 0, fmt.Errorf("can't align words: %w", err)
	}
	var res []*api.SpeechMark
	for i, w := range originalWords {
		if aligned[i] == -1 || aligned[i] >= len(maps) {
			continue
		}
		md := maps[aligned[i]]
		to := getLastWordTo(aligned, i, maps, data.SampleRate, md.part.Step)
		sm := &api.SpeechMark{
			Value:        w,
			Type:         api.SpeechMarkTypeWord,
			TimeInMillis: (from + md.start + utils.ToDuration(md.pw.SynthesizedPos.From+md.shift, data.SampleRate, md.part.Step)).Milliseconds(),
			Duration: (getLastWordTo(aligned, i, maps, data.SampleRate, md.part.Step) -
				(utils.ToDuration(md.pw.SynthesizedPos.From+md.shift, data.SampleRate, md.part.Step) + md.start)).Milliseconds(),
		}
		goapp.Log.Debug().Msgf("Word: %s, from: %d, to: %d, shift: %d, start: %d, from %d, res: %d-%d (%d)",
			w, md.pw.SynthesizedPos.From, to.Milliseconds(), md.shift, md.start.Milliseconds(), from.Milliseconds(), sm.TimeInMillis, sm.TimeInMillis+sm.Duration, sm.Duration)
		res = append(res, sm)
	}
	return res, calcDuration(data.Parts), nil
}

func calcDuration(tTSDataPart []*TTSDataPart) time.Duration {
	res := time.Duration(0)
	for _, p := range tTSDataPart {
		if p.AudioDurations != nil {
			res += p.AudioDurations.Duration
		}
	}
	return res
}

func getLastWordTo(aligned []int, i int, maps []*wordMapData, sampleRate uint32, step int) time.Duration {
	c := aligned[i]
	var upTo int
	if i < len(aligned)-1 {
		upTo = aligned[i+1]
	} else {
		upTo = len(maps)
	}

	res := c
	for i := c; i < upTo; i++ {
		if maps[i].pw.Tagged.IsWord() {
			res = i
		}
	}
	mw := maps[res]
	return utils.ToDuration(mw.pw.SynthesizedPos.To+mw.shift, sampleRate, step) + mw.start
}

func collectWords(parts []*TTSDataPart) ([]string, []*wordMapData) {
	var words []string
	var maps []*wordMapData
	startAt := time.Duration(0)
	for _, p := range parts {
		for _, w := range p.Words {
			if w.Tagged.IsWord() {
				words = append(words, w.Tagged.Word)
				maps = append(maps,
					&wordMapData{
						start: startAt,
						shift: p.AudioDurations.Shift,
						pw:    w,
						part:  p,
					})
			} else if w.Tagged.Separator != "" && len(words) > 0 {
				words[len(words)-1] += w.Tagged.Separator
			}
		}
		startAt += p.AudioDurations.Duration
	}
	return words, maps
}

func tryCustomCode(data *TTSData) {
	if strings.HasPrefix(data.OriginalText, "##AM:") {
		data.OriginalText = data.OriginalText[len("##AM:"):]
		data.Cfg.JustAM = true
		goapp.Log.Info().Msgf("Start from AM")
	}
}

func mapTranscribed(data *TTSData) string {
	res := strings.Builder{}
	write_func := func(s string) {
		if res.Len() > 0 {
			res.WriteRune(' ')
		}
		res.WriteString(s)
	}
	for _, p := range data.Parts {
		write_func(p.TranscribedText)
	}
	for _, sp := range data.SSMLParts {
		for _, p := range sp.Parts {
			write_func(p.TranscribedText)
		}
	}
	return res.String()
}

func mapAccentedText(data *TTSData) (string, error) {
	res := strings.Builder{}
	for _, p := range data.Parts {
		for _, w := range p.Words {
			tgw := w.Tagged
			if tgw.Space {
				res.WriteString(" ")
			} else if tgw.Separator != "" {
				res.WriteString(tgw.Separator)
			} else if tgw.IsWord() {
				aw, err := accent.ToAccentString(tgw.Word, GetTranscriberAccent(w))
				if err != nil {
					return "", errors.Wrapf(err, "Can't mark accent for %s", tgw.Word)
				}
				res.WriteString(aw)
			}
		}
	}
	return res.String(), nil
}

// GetTranscriberAccent return accent from ProcessedWord
func GetTranscriberAccent(w *ProcessedWord) int {
	if w.AccentVariant != nil {
		res := w.AccentVariant.Accent
		if w.UserAccent > 0 {
			res = w.UserAccent
		} else if w.TextPart != nil && w.TextPart.Accented != "" { // was empty accent provided if it not set in w.UserAccent
			res = 0
		} else if w.Clitic.Type == CliticsCustom {
			res = w.Clitic.Accent
		} else if w.Clitic.Type == CliticsNone {
			res = 0
		}
		return res
	}
	return 0
}

// GetProcessorsInfo return info about processors for testing
func (mw *MainWorker) GetProcessorsInfo() string {
	return getInfo(mw.processors)
}

// GetSSMLProcessorsInfo return info about processors for testing
func (mw *MainWorker) GetSSMLProcessorsInfo() string {
	return getInfo(mw.ssmlProcessors)
}

func getInfo(processors []Processor) string {
	res := strings.Builder{}
	nl := ""
	for _, pr := range processors {
		pri, ok := pr.(interface {
			Info() string
		})
		if ok {
			res.WriteString(nl + pri.Info())
			nl = "\n"
		}
	}
	return res.String()
}
