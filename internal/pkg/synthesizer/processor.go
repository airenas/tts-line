package synthesizer

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/airenas/tts-line/internal/pkg/accent"
	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/airenas/tts-line/internal/pkg/utils/dtw"
	"github.com/airenas/tts-line/pkg/ssml"
)

// Processor interface
type Processor interface {
	Process(context.Context, *TTSData) error
}

// MainWorker does synthesis work
type MainWorker struct {
	processors      []Processor
	ssmlProcessors  []Processor
	AllowCustomCode bool
}

// Work is main method
func (mw *MainWorker) Work(ctx context.Context, input *api.TTSRequestConfig) (*api.Result, error) {
	data := &TTSData{}
	data.OriginalText = input.Text
	data.Input = input
	data.Cfg.Input = input
	data.Cfg.Type = SSMLNone
	data.Cfg.SpeedRate = input.Speed
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
			utils.LogData(ctx, "Error", getInputText(data), err)
			return nil, err
		}
		if err := mw.processAll(ctx, mw.ssmlProcessors, data); err != nil {
			return nil, err
		}
	} else {
		if err := mw.processAll(ctx, mw.processors, data); err != nil {
			return nil, err
		}
	}
	return mapResult(ctx, data)
}

func makeSSMLParts(input *api.TTSRequestConfig) ([]*TTSData, error) {
	var res []*TTSData
	var last *TTSData
	for _, p := range input.SSMLParts {
		switch pc := p.(type) {
		case *ssml.Text:
			if last != nil && last.Cfg.Type == SSMLText && last.Cfg.Voice == pc.Voice {
				last.OriginalTextParts = append(last.OriginalTextParts, makeTextParts(pc.Texts, pc.Prosodies)...)
			} else {
				data := &TTSData{}
				data.OriginalTextParts = makeTextParts(pc.Texts, pc.Prosodies)
				data.Input = input
				data.Cfg.Input = input
				// data.Cfg.Prosodies = pc.Prosodies
				data.Cfg.Voice = pc.Voice
				data.Cfg.Type = SSMLText
				data.RequestID = input.RequestID
				if input.RequestID == "" {
					data.RequestID = uuid.NewString()
				}
				res = append(res, data)
				last = data
			}
		case *ssml.Pause:
			if pc.Duration <= time.Millisecond*1000 { // try merge small pauses
				if last != nil && last.Cfg.Type == SSMLText && len(last.OriginalTextParts) > 0 {
					last.OriginalTextParts[len(last.OriginalTextParts)-1].PauseAfter += pc.Duration
					continue
				}
			}
			data := &TTSData{}
			data.Cfg.PauseDuration = pc.Duration
			data.Cfg.Type = SSMLPause
			res = append(res, data)
			last = data
		default:
			return nil, errors.Errorf("unknown SSML part type %T", pc)
		}
	}
	return res, nil
}

func makeTextParts(textPart []ssml.TextPart, prosodies []*ssml.Prosody) []*TTSTextPart {
	res := []*TTSTextPart{}
	for _, tp := range textPart {
		res = append(res, &TTSTextPart{Text: tp.Text, Accented: tp.Accented, Syllables: tp.Syllables, UserOEPal: tp.UserOEPal,
			Language:          tp.Language,
			InterpretAs:       tp.InterpretAs,
			InterpretAsDetail: tp.InterpretAsDetail,
			Prosodies:         prosodies,
		})
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

func (mw *MainWorker) processAll(ctx context.Context, processors []Processor, data *TTSData) error {
	for _, pr := range processors {
		err := pr.Process(ctx, data)
		if err != nil {
			utils.LogData(ctx, "Error", getInputText(data), err)
			return err
		}
	}
	return nil
}

func getInputText(data *TTSData) string {
	if data.Input == nil {
		return ""
	}
	return data.Input.Text
}

func mapResult(ctx context.Context, data *TTSData) (*api.Result, error) {
	res := &api.Result{}
	res.Audio = data.AudioMP3
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
		res.SpeechMarks, err = mapSpeechMarks(ctx, data)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

type wordMapData struct {
	pw *ProcessedWord
}

func mapSpeechMarks(ctx context.Context, data *TTSData) ([]*api.SpeechMark, error) {
	if len(data.SSMLParts) == 0 {
		res, err := mapSpeechMarksInt(ctx, data)
		return res, err
	}

	var res []*api.SpeechMark
	for _, p := range data.SSMLParts {
		pRes, err := mapSpeechMarksInt(ctx, p)
		if err != nil {
			return nil, err
		}
		res = append(res, pRes...)
	}
	return res, nil
}

func mapSpeechMarksInt(ctx context.Context, data *TTSData) ([]*api.SpeechMark, error) {
	text := strings.Join(data.Text, " ")
	if len(text) == 0 {
		return nil, nil
	}
	originalWords := strings.Fields(accent.ClearAccents(text))
	originalWords = dropPunctuation(originalWords)
	words, maps := collectWords(data.Parts)
	aligned, err := dtw.Align(ctx, originalWords, words)
	if err != nil {
		return nil, fmt.Errorf("can't align words: %w", err)
	}
	var res []*api.SpeechMark
	for i, w := range originalWords {
		if aligned[i] == -1 || aligned[i] >= len(maps) {
			continue
		}
		md := maps[aligned[i]]
		to := getLastWordTo(aligned, i, maps, data.Audio.SampleRate, data.Audio.BitsPerSample)
		at := utils.BytesToDuration(md.pw.AudioPos.From, data.Audio.SampleRate, data.Audio.BitsPerSample)
		sm := &api.SpeechMark{
			Value: w,
			Type:  api.SpeechMarkTypeWord,
			// TimeInMillis: (from + md.start + utils.ToDuration(md.pw.SynthesizedPos.From+md.shift, data.Audio.SampleRate, md.part.Step)).Milliseconds(),
			TimeInMillis: at.Milliseconds(),
			Duration:     (to - at).Milliseconds(),
		}
		goapp.Log.Debug().Msgf("Word: %s, from: %d, to: %d, res: %d-%d (%d)",
			w, md.pw.SynthesizedPos.From, to.Milliseconds(), sm.TimeInMillis, sm.TimeInMillis+sm.Duration, sm.Duration)
		res = append(res, sm)
	}
	return res, nil
}

func dropPunctuation(originalWords []string) []string {
	var res []string
	for _, w := range originalWords {
		nw := strings.TrimFunc(w, func(_r rune) bool {
			return !unicode.IsLetter(_r) && !unicode.IsDigit(_r)
		})
		if nw == "" {
			continue
		}
		res = append(res, nw)
	}
	return res
}

func getLastWordTo(aligned []int, i int, maps []*wordMapData, sampleRate uint32, bitsPerSample uint16) time.Duration {
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
	return utils.BytesToDuration(mw.pw.AudioPos.To, sampleRate, bitsPerSample)
}

func collectWords(parts []*TTSDataPart) ([]string, []*wordMapData) {
	var words []string
	var maps []*wordMapData
	for _, p := range parts {
		for _, w := range p.Words {
			if w.Tagged.IsWord() {
				words = append(words, w.Tagged.Word)
				maps = append(maps,
					&wordMapData{
						pw: w,
					})
			}
		}
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
