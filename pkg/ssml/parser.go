package ssml

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"time"
)

type startFunc func(xml.StartElement, *wrkData) error
type endFunc func(xml.EndElement, *wrkData) error

type wrkData struct {
	speakTagCount    int
	speakTagEndCount int
	lastTag          []string
	voiceFunc        func(string) (string, error)

	res     []Part
	cValues []*Text
}

var startFunctions map[string]startFunc
var endFunctions map[string]endFunc
var durationStrs map[string]time.Duration
var pDuration time.Duration

func init() {
	startFunctions = make(map[string]startFunc)
	endFunctions = make(map[string]endFunc)

	startFunctions["speak"] = startSpeak
	endFunctions["speak"] = endSpeak
	startFunctions["p"] = startP
	endFunctions["p"] = endP
	startFunctions["break"] = startBreak
	endFunctions["break"] = endBreak
	startFunctions["voice"] = startVoice
	endFunctions["voice"] = endVoice

	durationStrs = map[string]time.Duration{"none": 0, "x-weak": 250 * time.Millisecond,
		"weak": 500 * time.Millisecond, "medium": 750 * time.Millisecond,
		"strong": 1000 * time.Millisecond, "x-strong": 1250 * time.Millisecond}
	pDuration = durationStrs["x-strong"]
}

// Parse parses xml into synthesis structure
func Parse(r io.Reader, def *Text, voiceFunc func(string) (string, error)) ([]Part, error) {
	wrk := &wrkData{res: make([]Part, 0), cValues: []*Text{def}, voiceFunc: voiceFunc}
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
			f, ok := startFunctions[se.Name.Local]
			if !ok {
				return nil, &ErrParse{Pos: d.InputOffset(), Msg: fmt.Sprintf("unknown tag <%s>", se.Name.Local)}
			}
			if err := f(se, wrk); err != nil {
				return nil, &ErrParse{Pos: d.InputOffset(), Msg: err.Error()}
			}
			wrk.lastTag = append(wrk.lastTag, se.Name.Local)
		case xml.EndElement:
			f, ok := endFunctions[se.Name.Local]
			if !ok {
				return nil, &ErrParse{Pos: d.InputOffset(), Msg: fmt.Sprintf("unknown tag </%s>", se.Name.Local)}
			}
			if err := f(se, wrk); err != nil {
				return nil, &ErrParse{Pos: d.InputOffset(), Msg: err.Error()}
			}
			wrk.lastTag = wrk.lastTag[:len(wrk.lastTag)-1]
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

func makeTextPart(se xml.CharData, wrk *wrkData) error {
	if wrk.speakTagCount != 1 {
		return fmt.Errorf("no <speak>")
	}
	if len(wrk.lastTag) == 0 {
		return fmt.Errorf("data after </speak>")
	}
	lt := wrk.lastTag[len(wrk.lastTag)-1]
	if lt == "break" {
		return fmt.Errorf("data in <break>")
	}
	s := strings.TrimSpace(string(se))
	if s != "" {
		def := wrk.cValues[len(wrk.cValues)-1]
		wrk.res = append(wrk.res, &Text{Text: s, Voice: def.Voice, Speed: def.Speed})
	}
	return nil
}

func startSpeak(se xml.StartElement, wrk *wrkData) error {
	if wrk.speakTagCount != 0 {
		return fmt.Errorf("multiple <speak>")
	}
	wrk.speakTagCount++
	return nil
}

func endSpeak(se xml.EndElement, wrk *wrkData) error {
	wrk.speakTagEndCount++
	return nil
}

func startP(se xml.StartElement, wrk *wrkData) error {
	if wrk.speakTagCount != 1 {
		return fmt.Errorf("no <speak>")
	}
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
			return 0, fmt.Errorf("wrong duration '%s'", str)
		}
		return res, nil
	}
	return 0, fmt.Errorf("no duration")
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
	return nil
}

func startVoice(se xml.StartElement, wrk *wrkData) error {
	if wrk.speakTagCount != 1 {
		return fmt.Errorf("no <speak>")
	}
	v := getAttr(se, "name")
	if v == "" {
		return fmt.Errorf("no <voice> name")
	}
	var err error
	v, err = wrk.voiceFunc(v)
	if err != nil {
		return err
	}
	def := wrk.cValues[len(wrk.cValues)-1]
	wrk.cValues = append(wrk.cValues, &Text{Voice: v, Speed: def.Speed})
	return nil
}

func endVoice(se xml.EndElement, wrk *wrkData) error {
	if wrk.speakTagCount != 1 {
		return fmt.Errorf("no </speak>")
	}
	wrk.cValues = wrk.cValues[:len(wrk.cValues)-1]
	return nil
}
