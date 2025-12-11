package ssml

import (
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/airenas/tts-line/internal/pkg/accent"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/rs/zerolog/log"
)

type startFunc func(xml.StartElement, *wrkData) error
type endFunc func(xml.EndElement, *wrkData) error

type tagsData[T any] struct {
	values []T
}

func (t *tagsData[T]) peek() T {
	if len(t.values) == 0 {
		var zero T
		return zero
	}
	return t.values[len(t.values)-1]
}

func (t *tagsData[T]) count() int {
	return len(t.values)
}

func (t *tagsData[T]) pop() {
	if len(t.values) == 0 {
		return
	}
	t.values = t.values[:len(t.values)-1]
}

func (t *tagsData[T]) push(lang T) {
	t.values = append(t.values, lang)
}

type wrkData struct {
	speakTagCount    int
	speakTagEndCount int
	languages        tagsData[string]
	prosodies        tagsData[*Prosody]
	voices           tagsData[string]

	lastTag   []string
	voiceFunc func(string) (string, error)
	langFunc  func(string) (string, error)

	lastWAcc, lastWSyll, lastWUser string

	lastInterpretAs       InterpretAsType
	lastInterpretAsDetail InterpretAsDetailType

	lastText *Text

	res []Part
	// cValues []*Text
	emphasisCount int
}

func (w *wrkData) validateNew(key string) error {
	if w.allowedInside(key) {
		return nil
	}
	lt := w.getLastTag()
	if lt == "" {
		return fmt.Errorf("not allowed tag <%s> as the root element", key)
	}
	return fmt.Errorf("not allowed tag <%s> inside <%s>", key, lt)
}

func (w *wrkData) allowedInside(key string) bool {
	lastTag := w.getLastTag()
	if lastTagMap, ok := allowedInside[key]; ok {
		_, ok := lastTagMap[lastTag]
		return ok
	}
	return false
}

func (w *wrkData) getLastTag() string {
	if len(w.lastTag) == 0 {
		return ""
	}
	return w.lastTag[len(w.lastTag)-1]
}

var startFunctions map[string]startFunc
var endFunctions map[string]endFunc
var allowedInside map[string]map[string]struct{}
var durationStrs map[string]time.Duration
var rateStrs map[string]float64
var volumeStrs map[string]float64
var pitchStrs map[string]float64
var emphasisLevels map[string]EmphasisType
var interpretAsTypes map[string]InterpretAsType
var interpretAsDetailTypes map[string]InterpretAsDetailType
var pDuration time.Duration

const (
	TagSpeak    = "speak"
	TagP        = "p"
	TagBreak    = "break"
	TagVoice    = "voice"
	TagProsody  = "prosody"
	TagWord     = "intelektika:w"
	TagLang     = "lang"
	TagEmphasis = "emphasis"
	TagSayAs    = "say-as"
)

func init() {
	startFunctions = make(map[string]startFunc)
	endFunctions = make(map[string]endFunc)
	allowedInside = make(map[string]map[string]struct{})

	startFunctions[TagSpeak] = startSpeak
	endFunctions[TagSpeak] = endSpeak
	allowedInside[TagSpeak] = makeTagMap("")

	startFunctions[TagP] = startP
	endFunctions[TagP] = endP
	allowedInside[TagP] = makeTagMap(TagSpeak, TagEmphasis, TagProsody, TagVoice, TagP, TagLang)

	startFunctions[TagBreak] = startBreak
	endFunctions[TagBreak] = endBreak
	allowedInside[TagBreak] = makeTagMap(TagSpeak, TagEmphasis, TagProsody, TagVoice, TagP, TagLang)

	startFunctions[TagVoice] = startVoice
	endFunctions[TagVoice] = endVoice
	allowedInside[TagVoice] = makeTagMap(TagSpeak, TagEmphasis, TagProsody, TagVoice, TagP, TagLang)

	startFunctions[TagProsody] = startProsody
	endFunctions[TagProsody] = endProsody
	allowedInside[TagProsody] = makeTagMap(TagSpeak, TagEmphasis, TagProsody, TagVoice, TagP, TagLang)

	startFunctions[TagWord] = startWord
	endFunctions[TagWord] = endWord
	allowedInside[TagWord] = makeTagMap(TagSpeak, TagEmphasis, TagProsody, TagVoice, TagP, TagLang)

	startFunctions[TagLang] = startLang
	endFunctions[TagLang] = endLang
	allowedInside[TagLang] = makeTagMap(TagSpeak, TagEmphasis, TagProsody, TagVoice, TagP, TagLang)

	startFunctions[TagEmphasis] = startEmphasis
	endFunctions[TagEmphasis] = endEmphasis
	allowedInside[TagEmphasis] = makeTagMap(TagSpeak, TagEmphasis, TagProsody, TagVoice, TagP, TagLang)

	startFunctions[TagSayAs] = startSayAs
	endFunctions[TagSayAs] = endSayAs
	allowedInside[TagSayAs] = makeTagMap(TagSpeak, TagEmphasis, TagProsody, TagVoice, TagP, TagLang)

	durationStrs = map[string]time.Duration{"none": 0, "x-weak": 250 * time.Millisecond,
		"weak": 500 * time.Millisecond, "medium": 750 * time.Millisecond,
		"strong": 1000 * time.Millisecond, "x-strong": 1250 * time.Millisecond}
	pDuration = durationStrs["x-strong"]

	rateStrs = map[string]float64{"x-slow": 2, "slow": 1.5, "medium": 1, "fast": .75, "x-fast": .5, "default": 1}
	volumeStrs = map[string]float64{"silent": MinVolumeChange, "x-soft": -6, "soft": -3, "medium": 0, "loud": 3, "x-loud": 6, "default": 0}
	pitchStrs = map[string]float64{"x-low": 0.55, "low": 0.8, "medium": 1, "high": 1.2, "x-high": 1.45, "default": 1}
	emphasisLevels = map[string]EmphasisType{"reduced": EmphasisTypeReduced, "none": EmphasisTypeNone,
		"moderate": EmphasisTypeModerate, "strong": EmphasisTypeStrong}
	interpretAsTypes = map[string]InterpretAsType{InterpretAsTypeCharacters.String(): InterpretAsTypeCharacters}
	interpretAsDetailTypes = map[string]InterpretAsDetailType{InterpretAsDetailTypeReadSymbols.String(): InterpretAsDetailTypeReadSymbols}
}

func makeTagMap(s ...string) map[string]struct{} {
	res := make(map[string]struct{})
	for _, v := range s {
		res[v] = struct{}{}
	}
	return res
}

// Parse parses xml into synthesis structure
func Parse(r io.Reader, def *Text, voiceFunc func(string) (string, error)) ([]Part, error) {
	wrk := &wrkData{res: make([]Part, 0),
		//cValues: []*Text{def},
		voiceFunc: voiceFunc,
		langFunc:  checkLanguage,
	}
	wrk.voices.push(def.Voice)

	d := xml.NewDecoder(r)

	for {
		// Read tokens from the XML document in a stream.
		t, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("ssml: %v", err)
		}
		if t == nil {
			break
		}
		// Inspect the type of the token just read.
		switch se := t.(type) {
		case xml.StartElement:
			key := getXMLKey(se.Name)
			f, ok := startFunctions[key]
			if !ok {
				return nil, &ErrParse{Pos: d.InputOffset(), Msg: fmt.Sprintf("unknown tag <%s>", key)}
			}
			if err := wrk.validateNew(key); err != nil {
				return nil, &ErrParse{Pos: d.InputOffset(), Msg: err.Error()}
			}
			if err := f(se, wrk); err != nil {
				return nil, &ErrParse{Pos: d.InputOffset(), Msg: err.Error()}
			}
			wrk.lastTag = append(wrk.lastTag, key)
		case xml.EndElement:
			key := getXMLKey(se.Name)
			f, ok := endFunctions[key]
			if !ok {
				return nil, &ErrParse{Pos: d.InputOffset(), Msg: fmt.Sprintf("unknown tag </%s>", key)}
			}
			if err := f(se, wrk); err != nil {
				return nil, &ErrParse{Pos: d.InputOffset(), Msg: err.Error()}
			}
			l := len(wrk.lastTag) - 1
			wrk.lastTag[l] = ""
			wrk.lastTag = wrk.lastTag[:l]
		case xml.CharData:
			err := makeTextPart(se, wrk)
			if err != nil {
				return nil, &ErrParse{Pos: d.InputOffset(), Msg: err.Error()}
			}
		case xml.Comment:
		case xml.ProcInst:
		case xml.Directive:
		default:
			return nil, &ErrParse{Pos: d.InputOffset(), Msg: fmt.Sprintf("unknown element %v", se)}
		}
	}
	if wrk.speakTagEndCount != 1 {
		return nil, &ErrParse{Pos: d.InputOffset(), Msg: "no </speak>"}
	}
	return wrk.res, nil
}

func getXMLKey(name xml.Name) string {
	if name.Space != "" {
		return name.Space + ":" + name.Local
	}
	return name.Local
}

func makeTextPart(se xml.CharData, wrk *wrkData) error {
	s := strings.TrimSpace(string(se))
	if s != "" {
		// def := wrk.cValues[len(wrk.cValues)-1]
		if wrk.speakTagCount != 1 {
			return fmt.Errorf("no <speak>")
		}
		if len(wrk.lastTag) == 0 {
			return fmt.Errorf("data after </speak>")
		}
		lt := wrk.getLastTag()
		if lt == TagBreak {
			return fmt.Errorf("data in <break>")
		}
		tp := TextPart{Text: s, Language: wrk.languages.peek()}
		if wrk.lastWAcc != "" {
			tp.Accented = wrk.lastWAcc
			wrk.lastWAcc = ""
		}
		tp.Syllables = wrk.lastWSyll
		tp.UserOEPal = wrk.lastWUser
		tp.InterpretAs = wrk.lastInterpretAs
		tp.InterpretAsDetail = wrk.lastInterpretAsDetail
		if wrk.lastText != nil {
			wrk.lastText.Texts = append(wrk.lastText.Texts, tp)
		} else {
			wrk.lastText = &Text{Voice: wrk.voices.peek(), Texts: []TextPart{tp}, Prosodies: utils.SlicesCopy(wrk.prosodies.values)}
			wrk.res = append(wrk.res, wrk.lastText)
		}
	}
	return nil
}

func startSpeak(se xml.StartElement, wrk *wrkData) error {
	if wrk.speakTagCount != 0 {
		return fmt.Errorf("wrong <speak>")
	}
	wrk.speakTagCount++

	lang := getAttr(se, "lang")
	var err error
	lang, err = wrk.langFunc(lang)
	if err != nil {
		return err
	}
	wrk.languages.push(lang)

	return nil
}

func endSpeak(se xml.EndElement, wrk *wrkData) error {
	wrk.speakTagEndCount++
	if wrk.languages.count() < 1 {
		return fmt.Errorf("no <speak>")
	}
	wrk.languages.pop()
	return nil
}

func startP(se xml.StartElement, wrk *wrkData) error {
	if wrk.speakTagCount != 1 {
		return fmt.Errorf("no <speak>")
	}
	wrk.lastText = nil
	wrk.res = append(wrk.res, &Pause{Duration: pDuration})
	return nil
}

func endP(se xml.EndElement, wrk *wrkData) error {
	if wrk.speakTagCount != 1 {
		return fmt.Errorf("no </speak>")
	}
	if len(wrk.res) > 0 && !IsPause(wrk.res[len(wrk.res)-1]) {
		wrk.res = append(wrk.res, &Pause{Duration: pDuration})
	}
	wrk.lastText = nil
	return nil
}

func startBreak(se xml.StartElement, wrk *wrkData) error {
	if wrk.speakTagCount != 1 {
		return fmt.Errorf("no <speak>")
	}
	d, err := getDuration(getAttr(se, "time"), getAttr(se, "strength"))
	if err != nil {
		return err
	}
	wrk.lastText = nil
	wrk.res = append(wrk.res, &Pause{Duration: d, IsBreak: true})
	return nil
}

func getDuration(tm, str string) (time.Duration, error) {
	if tm != "" {
		res, err := time.ParseDuration(tm)
		if err != nil {
			return 0, err
		}
		return res, nil
	} else if str != "" {
		res, ok := durationStrs[str]
		if !ok {
			return 0, fmt.Errorf("wrong <break>:strength '%s'", str)
		}
		return res, nil
	}
	return 0, fmt.Errorf("no <break>:time|strength")
}

func getAttr(se xml.StartElement, s string) string {
	for _, n := range se.Attr {
		if n.Name.Local == s {
			return n.Value
		}
	}
	return ""
}

func endBreak(se xml.EndElement, wrk *wrkData) error {
	if wrk.speakTagCount != 1 {
		return fmt.Errorf("no </speak>")
	}
	wrk.lastText = nil
	return nil
}

func startVoice(se xml.StartElement, wrk *wrkData) error {
	if wrk.speakTagCount != 1 {
		return fmt.Errorf("no <speak>")
	}
	v := getAttr(se, "name")
	if v == "" {
		return fmt.Errorf("no <voice>:name")
	}
	var err error
	v, err = wrk.voiceFunc(v)
	if err != nil {
		return err
	}
	wrk.lastText = nil
	wrk.voices.push(v)
	return nil
}

func endVoice(se xml.EndElement, wrk *wrkData) error {
	if wrk.speakTagCount != 1 {
		return fmt.Errorf("no </speak>")
	}
	if wrk.voices.count() < 1 {
		return fmt.Errorf("no <voice>")
	}
	wrk.voices.pop()
	wrk.lastText = nil
	return nil
}

func startProsody(se xml.StartElement, wrk *wrkData) error {
	if wrk.speakTagCount != 1 {
		return fmt.Errorf("no <speak>")
	}
	was := false
	var err error

	sp := float64(1)
	r := getAttr(se, "rate")
	if r != "" {
		sp, err = getSpeed(r)
		if err != nil {
			return err
		}
		was = true
	}
	volumeChange := float64(0)
	r = getAttr(se, "volume")
	if r != "" {
		volumeChange, err = getVolume(r)
		if err != nil {
			return err
		}
		was = true
	}
	var pitchChange PitchChange

	r = getAttr(se, "pitch")
	if r != "" {
		pitchChange, err = getPitch(r)
		if err != nil {
			return err
		}
		was = true
	}
	if !was {
		return fmt.Errorf("no <prosody>:rate or <prosody>:volume or <prosody>:pitch")
	}
	wrk.prosodies.push(&Prosody{Rate: sp, Volume: volumeChange, Pitch: pitchChange})
	wrk.lastText = nil
	// def := wrk.cValues[len(wrk.cValues)-1]
	// newTextPart := makeInternalText(def, &Text{Speed: sp})
	// newTextPart.Prosodies = slices.Clone(wrk.prosodies.values)
	// wrk.cValues = append(wrk.cValues, newTextPart)
	return nil
}

func endProsody(se xml.EndElement, wrk *wrkData) error {
	if wrk.speakTagCount != 1 {
		return fmt.Errorf("no </speak>")
	}
	if wrk.prosodies.count() < 1 {
		return fmt.Errorf("no <prosody>")
	}
	wrk.prosodies.pop()
	wrk.lastText = nil
	return nil
}

func startWord(se xml.StartElement, wrk *wrkData) error {
	if wrk.speakTagCount != 1 {
		return fmt.Errorf("no <speak>")
	}
	lt := wrk.lastTag[len(wrk.lastTag)-1]
	if lt != "speak" && lt != "p" && lt != "prosody" && lt != "voice" {
		return fmt.Errorf("<intelektika:w> not allowed inside <%s>", lt)
	}
	a := getAttr(se, "acc")
	if !okAccentedWord(a) {
		return fmt.Errorf("wrong <intelektika:w>:acc='%s'", a)
	}
	wrk.lastWAcc = a
	wrk.lastWSyll = getAttr(se, "syll")
	if wrk.lastWSyll != "" && accent.ClearAccents(wrk.lastWAcc) != clearSylls(wrk.lastWSyll) {
		return fmt.Errorf("wrong syllables <intelektika:w>:acc='%s' syll='%s'", wrk.lastWAcc, wrk.lastWSyll)
	}
	wrk.lastWUser = getAttr(se, "user")
	if wrk.lastWUser != "" && !strings.EqualFold(accent.ClearAccents(wrk.lastWAcc), clearUserOE(wrk.lastWUser)) {
		return fmt.Errorf("wrong OE model <intelektika:w>:acc='%s' user='%s'", wrk.lastWAcc, wrk.lastWUser)
	}
	return nil
}

func endWord(se xml.EndElement, wrk *wrkData) error {
	if wrk.speakTagCount != 1 {
		return fmt.Errorf("no </speak>")
	}
	if wrk.lastWAcc != "" {
		return fmt.Errorf("no word in <intelektika:w>")
	}
	return nil
}

func startLang(se xml.StartElement, wrk *wrkData) error {
	if wrk.speakTagCount != 1 {
		return fmt.Errorf("no <speak>")
	}
	lang := getAttr(se, "lang")
	if lang == "" {
		return fmt.Errorf("no <lang>:lang")
	}
	var err error
	lang, err = wrk.langFunc(lang)
	if err != nil {
		return err
	}
	wrk.languages.push(lang)
	return nil
}

func endLang(se xml.EndElement, wrk *wrkData) error {
	if wrk.speakTagCount != 1 {
		return fmt.Errorf("no </speak>")
	}
	if wrk.languages.count() < 1 {
		return fmt.Errorf("no </lang>")
	}
	wrk.languages.pop()
	return nil
}

func startEmphasis(se xml.StartElement, wrk *wrkData) error {
	if wrk.speakTagCount != 1 {
		return fmt.Errorf("no <speak>")
	}
	levelStr := getAttr(se, "level")
	if levelStr == "" {
		levelStr = "moderate"
	}
	level, ok := emphasisLevels[levelStr]
	if !ok {
		return fmt.Errorf("wrong <emphasis>:level='%s'", levelStr)
	}

	wrk.emphasisCount++
	prosody := &Prosody{Emphasis: level, Rate: 1, ID: wrk.emphasisCount}
	wrk.prosodies.push(prosody)
	wrk.lastText = nil
	return nil
}

func endEmphasis(se xml.EndElement, wrk *wrkData) error {
	if wrk.speakTagCount != 1 {
		return fmt.Errorf("no </speak>")
	}
	if wrk.prosodies.count() < 1 {
		return fmt.Errorf("no <emphasis>")
	}
	wrk.prosodies.pop()
	wrk.lastText = nil
	return nil
}

func startSayAs(se xml.StartElement, wrk *wrkData) error {
	interpretAsStr := getAttr(se, "interpret-as")
	interpretAs, ok := interpretAsTypes[interpretAsStr]
	if !ok {
		return fmt.Errorf("wrong <say-as>:interpret-as='%s'", interpretAsStr)
	}
	detailStr := getAttr(se, "detail")
	detail, ok := interpretAsDetailTypes[detailStr]
	if !ok && detailStr != "" {
		log.Warn().Str("value", detailStr).Msg("Unknown value in <say-as>:detail")
		return fmt.Errorf("wrong <say-as>:detail='%s'", detailStr)
	}
	wrk.lastInterpretAs = interpretAs
	wrk.lastInterpretAsDetail = detail
	return nil
}

func endSayAs(se xml.EndElement, wrk *wrkData) error {
	wrk.lastInterpretAs = InterpretAsTypeUnset
	wrk.lastInterpretAsDetail = InterpretAsDetailTypeUnset
	return nil
}

func clearUserOE(s string) string {
	r := strings.NewReplacer("'", "", "*", "")
	return r.Replace(s)
}

func clearSylls(s string) string {
	return strings.ReplaceAll(s, "-", "")
}

func okAccentedWord(a string) bool {
	return a != "" && len(a) < 50 && accent.IsWordOrWithAccent(a)
}

func getSpeed(str string) (float64, error) {
	if str == "" {
		return 0, fmt.Errorf("no rate")
	}
	if strings.HasSuffix(str, "%") {
		v, err := strconv.ParseFloat(str[:len(str)-1], 64)
		if err != nil {
			return 0, fmt.Errorf("wrong rate value '%s': %v", str, err)
		}
		return parseRatePercents(v), nil
	}
	res, ok := rateStrs[str]
	if !ok {
		return 0, fmt.Errorf("wrong rate value '%s'", str)
	}
	return res, nil
}

func getVolume(str string) (float64, error) {
	if str == "" {
		return 0, fmt.Errorf("no volume")
	}
	if strings.HasSuffix(str, "dB") && (strings.HasPrefix(str, "+") || strings.HasPrefix(str, "-")) {
		v, err := strconv.ParseFloat(str[:len(str)-2], 64)
		if err != nil {
			return 0, fmt.Errorf("wrong volume value '%s': %v", str, err)
		}
		return v, nil
	}
	res, ok := volumeStrs[str]
	if !ok {
		return 0, fmt.Errorf("wrong volume value '%s'", str)
	}
	return res, nil
}

func getPitch(str string) (PitchChange, error) {
	if str == "" {
		return PitchChange{}, fmt.Errorf("no pitch")
	}
	if strings.HasSuffix(str, "Hz") && (strings.HasPrefix(str, "+") || strings.HasPrefix(str, "-")) {
		v, err := strconv.ParseFloat(str[:len(str)-2], 64)
		if err != nil {
			return PitchChange{}, fmt.Errorf("wrong pitch value '%s': %v", str, err)
		}
		return PitchChange{Kind: PitchChangeHertz, Value: v}, nil
	}
	if strings.HasSuffix(str, "st") && (strings.HasPrefix(str, "+") || strings.HasPrefix(str, "-")) {
		v, err := strconv.ParseFloat(str[:len(str)-2], 64)
		if err != nil {
			return PitchChange{}, fmt.Errorf("wrong pitch value '%s': %v", str, err)
		}
		return PitchChange{Kind: PitchChangeSemitone, Value: v}, nil
	}
	if strings.HasSuffix(str, "%") && (strings.HasPrefix(str, "+") || strings.HasPrefix(str, "-")) {
		v, err := strconv.ParseFloat(str[:len(str)-1], 64)
		if err != nil {
			return PitchChange{}, fmt.Errorf("wrong pitch value '%s': %v", str, err)
		}
		return PitchChange{Kind: PitchChangeMultiplier, Value: percentToMultiplier(v)}, nil
	}
	res, ok := pitchStrs[str]
	if !ok {
		return PitchChange{}, fmt.Errorf("wrong pitch value '%s'", str)
	}
	return PitchChange{Kind: PitchChangeMultiplier, Value: res}, nil
}

func parseRatePercents(v float64) float64 {
	if v > 200 {
		v = 200
	} else if v < 50 {
		v = 50
	}
	if v < 100 {
		return 1 + (100-v)/50
	}
	return 1 - (v-100)/200
}

func percentToMultiplier(v float64) float64 {
	return 1 + v/100.0
}

func checkLanguage(lang string) (string, error) {
	res := strings.ToLower(strings.TrimSpace(lang))
	if res == "" || res == "lt" {
		return "", nil
	}
	if len(res) == 2 {
		return res, nil
	}
	return "", fmt.Errorf("unsupported language '%s'", lang)
}
