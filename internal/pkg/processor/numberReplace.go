package processor

import (
	"fmt"
	"strings"
	"time"

	"github.com/airenas/go-app/pkg/goapp"
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
		goapp.Log.Info("Skip numberReplace")
		return nil
	}
	res := ""
	err := p.httpWrap.InvokeText(clearAccents(strings.Join(data.Text, "")), &res)
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
	err := p.httpWrap.InvokeText(clearAccents(strings.Join(data.Text, "")), &res)
	if err != nil {
		return err
	}
	data.TextWithNumbers, err = mapAccentsBack(res, data.Text)
	return err
}

func mapAccentsBack(new string, origArr []string) ([]string, error) {
	orig := strings.Join(origArr, "")
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
	// if len(accWrds) == 0 {
	// 	return new, nil
	// }

	alignIDs, err := align(ocStrs, nStrs, 20)
	if err != nil {
		// try increase align buffer, maybe it will help
		goapp.Log.Info("increase align size")
		alignIDs, err = align(ocStrs, nStrs, 40)
		if err != nil {
			return nil, errors.Wrapf(err, "can't align")
		}
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
	wi := 0
	for _, s := range origArr{
		l := strings.Count(s, " ")
		nID := len(alignIDs)
		if nID > (l + wi){
			nID = alignIDs[l + wi]
		} 
		if nID == -1 {
			return nil, errors.Errorf("no word alignment for %v", s)
		}
		nID += 1
		res = append(res, strings.Join(nStrs[wi:nID], " "))
		wi += l
	}
	return res, nil
}

func align(oStrs []string, nStrs []string, step int) ([]int, error) {
	res := make([]int, len(oStrs))
	i, j := 0, 0
	for i < len(oStrs) {
		partlyAll := doPartlyAlign(oStrs[i:min(i+step, len(oStrs))], nStrs[j:min(j+step, len(nStrs))], step)
		for pi, v := range partlyAll {
			if v > -1 {
				res[i+pi] = partlyAll[pi] + j
			} else {
				res[i+pi] = -1
			}
		}
		if min(i+step, len(oStrs)) == len(oStrs) && min(j+step, len(nStrs)) == len(nStrs) {
			break
		}
		nextI, err := findLastConsequtiveAlign(partlyAll, oStrs[i:], nStrs[j:])
		if err != nil {
			return nil, errors.Wrapf(err, "can't find next steps starting at %d, %d", i, j)
		}
		i = i + nextI
		j = j + partlyAll[nextI]
	}
	return res, nil
}

type moveType byte

const (
	start moveType = iota
	left
	top
	corner
)

func doPartlyAlign(s1 []string, s2 []string, step int) []int {
	ind := func(i1, i2 int) int {
		return i1*step + i2
	}
	l1, l2 := len(s1), len(s2)
	h := make([]byte, step*step)
	hb := make([]moveType, step*step)
	// calc h and h backtrace matrices
	for i1 := 0; i1 < l1; i1++ {
		for i2 := 0; i2 < l2; i2++ {
			eq := s1[i1] == s2[i2]
			eqv := byte(1)
			if eq {
				eqv = 0
			}
			if i1 == 0 && i2 == 0 {
				h[ind(i1, i2)] = byte(eqv)
				hb[ind(i1, i2)] = start
			} else if i1 == 0 {
				h[ind(i1, i2)] = h[ind(i1, i2-1)] + 1
				hb[ind(i1, i2)] = left
			} else if i2 == 0 {
				h[ind(i1, i2)] = h[ind(i1-1, i2)] + 1
				hb[ind(i1, i2)] = top
			} else {
				cv := h[ind(i1-1, i2-1)] + eqv
				lv := h[ind(i1, i2-1)] + 1
				tv := h[ind(i1-1, i2)] + 1

				if cv <= tv && cv <= lv {
					h[ind(i1, i2)] = cv
					hb[ind(i1, i2)] = corner
					// prefer left or top move if not eq
					if !eq && (tv <= cv || lv <= cv) {
						if lv <= tv {
							hb[ind(i1, i2)] = left
						} else {
							hb[ind(i1, i2)] = top
						}
					}
				} else if lv <= tv {
					h[ind(i1, i2)] = lv
					hb[ind(i1, i2)] = left
				} else {
					h[ind(i1, i2)] = tv
					hb[ind(i1, i2)] = top
				}
			}
		}
	}
	res := make([]int, l1)
	i2 := l2 - 1
	for i1 := l1 - 1; i1 >= 0; i1-- {
		v := hb[ind(i1, i2)]
		if v == corner {
			res[i1] = i2
			i2--
		} else if v == left {
			for v == left && i2 > 0 {
				i2--
				v = hb[ind(i1, i2)]
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

// findLastConsequtiveAlign find the last two words matching in a row
func findLastConsequtiveAlign(partlyAll []int, oStrs []string, nStrs []string) (int, error) {

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
	return -1, errors.New("no consequtive match found in partly aligned sequences")
}

func (p *ssmlNumberReplace) skip(data *synthesizer.TTSData) bool {
	return data.Cfg.JustAM
}

// Info return info about processor
func (p *ssmlNumberReplace) Info() string {
	return fmt.Sprintf("SSMLNumberReplace(%s)", utils.RetrieveInfo(p.httpWrap))
}
