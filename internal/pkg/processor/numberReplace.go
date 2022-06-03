package processor

import (
	"strings"
	"time"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/pkg/errors"
)

//HTTPInvoker makes http call
type HTTPInvoker interface {
	InvokeText(string, interface{}) error
}

type numberReplace struct {
	httpWrap HTTPInvoker
}

//NewNumberReplace creates new processor
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
		goapp.Log.Info("Skip numberReplace")
		return nil
	}
	return p.httpWrap.InvokeText(data.Text, &data.TextWithNumbers)
}

func (p *numberReplace) skip(data *synthesizer.TTSData) bool {
	return data.Cfg.JustAM
}

type ssmlNumberReplace struct {
	httpWrap HTTPInvoker
}

// NewSSMLNumberReplace creates new processor
func NewSSMLNumberReplace(urlStr string) (synthesizer.Processor, error) {
	res := &ssmlNumberReplace{}
	var err error
	res.httpWrap, err = newHTTPWrapBackoff(urlStr, time.Second*10)

	if err != nil {
		return nil, errors.Wrap(err, "can't init http client")
	}
	return res, nil
}

func (p *ssmlNumberReplace) Process(data *synthesizer.TTSData) error {
	if p.skip(data) {
		goapp.Log.Info("Skip numberReplace")
		return nil
	}
	res := ""
	err := p.httpWrap.InvokeText(clearAccents(data.Text), &res)
	if err != nil {
		return err
	}
	data.TextWithNumbers, err = mapAccentsBack(res, data.Text)
	return err
}

func mapAccentsBack(new, orig string) (string, error) {
	oStrs := strings.Split(orig, " ")
	nStrs := strings.Split(new, " ")
	accWrds := map[int]bool{}
	ocStrs := make([]string, len(oStrs))
	for i, s := range oStrs {
		ocStrs[i] = clearAccents(s)
		if ocStrs[i] != s {
			accWrds[i] = true
		}
	}
	if len(accWrds) == 0 {
		return new, nil
	}

	allignIDs, err := allign(ocStrs, nStrs)
	if err != nil {
		return "", errors.Wrapf(err, "can't allign")
	}
	for k, _ := range accWrds {
		nID := allignIDs[k]
		if nID == -1 {
			return "", errors.Errorf("no word allignment for %s, ID: %d", oStrs[k], k)
		}
		nStrs[nID] = oStrs[k]
	}
	return strings.Join(nStrs, " "), nil
}

const partlyStep = 20

func allign(oStrs []string, nStrs []string) ([]int, error) {
	res := make([]int, len(oStrs))
	i, j := 0, 0
	for i < len(oStrs) {
		partlyAll := doPartlyAllign(oStrs[i:min(i+partlyStep, len(oStrs))], nStrs[j:min(j+partlyStep, len(nStrs))])
		for pi, v := range partlyAll {
			if v > -1 {
				res[i+pi] = partlyAll[pi] + j
			} else {
				res[i+pi] = -1
			}
		}
		if min(i+partlyStep, len(oStrs)) == len(oStrs) && min(j+partlyStep, len(nStrs)) == len(nStrs) {
			break
		}
		nextI, err := findLastConsequtiveAllign(partlyAll, oStrs[i:], nStrs[j:])
		if err != nil {
			return nil, errors.Wrapf(err, "can't find next steps starting at %d, %d", i, j)
		}
		i = i + nextI
		j = j + partlyAll[nextI]
	}
	return res, nil
}

const (
	start byte = iota
	left
	top
	corner
)

func doPartlyAllign(s1 []string, s2 []string) []int {
	l1, l2 := len(s1), len(s2)
	h := [partlyStep][partlyStep]byte{}
	hb := [partlyStep][partlyStep]byte{}
	// calc h and h backtrace matrices
	for i1 := 0; i1 < l1; i1++ {
		for i2 := 0; i2 < l2; i2++ {
			eq := s1[i1] == s2[i2]
			eqv := byte(1)
			if eq {
				eqv = 0
			}
			if i1 == 0 && i2 == 0 {
				h[i1][i2] = byte(eqv)
				hb[i1][i2] = start
			} else if i1 == 0 {
				h[i1][i2] = h[i1][i2-1] + 1
				hb[i1][i2] = left
			} else if i2 == 0 {
				h[i1][i2] = h[i1-1][i2] + 1
				hb[i1][i2] = top
			} else {
				cv := h[i1-1][i2-1] + eqv
				lv := h[i1][i2-1] + 1
				tv := h[i1-1][i2] + 1

				if cv <= tv && cv <= lv {
					h[i1][i2] = cv
					hb[i1][i2] = corner
					// prefer left ot top move qif not eq
					if !eq && (tv <= cv || lv <= cv) {
						if lv <= tv {
							hb[i1][i2] = left
						} else {
							hb[i1][i2] = top
						}
					}
				} else if lv <= tv {
					h[i1][i2] = lv
					hb[i1][i2] = left
				} else {
					h[i1][i2] = tv
					hb[i1][i2] = top
				}
			}
		}
	}
	res := make([]int, l1)
	i2 := l2 - 1
	for i1 := l1 - 1; i1 >= 0; i1-- {
		v := hb[i1][i2]
		if v == corner {
			res[i1] = i2
			i2--
		} else if v == left {
			for v == left && i2 > 0 {
				i2--
				v = hb[i1][i2]
			}
			res[i1] = i2
			i2--
		} else if v == top {
			res[i1] = -1
		} else { // start
			res[i1] = 0
		}
	}
	return res
}

func findLastConsequtiveAllign(partlyAll []int, oStrs []string, nStrs []string) (int, error) {
	// find the last two words matching in a row
	for i := len(partlyAll) - 1; i > 1; i-- {
		v := partlyAll[i]
		if v > -1 && v-1 == partlyAll[i-1] {
			if oStrs[i-1] != nStrs[partlyAll[i-1]] {
				i--
			} else if oStrs[i] == nStrs[partlyAll[i]] {
				return i, nil
			}
		}
	}
	return -1, errors.New("no consequtive match found in partly alligned sequences")
}

func (p *ssmlNumberReplace) skip(data *synthesizer.TTSData) bool {
	return data.Cfg.JustAM
}
