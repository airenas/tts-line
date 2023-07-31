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
		goapp.Log.Info("Skip numberReplace")
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
	err := p.httpWrap.InvokeText(accent.ClearAccents(strings.Join(data.Text, "")), &res)
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

	var alignIDs []int
	var err error
	for i, alignBuffer := range [...]int{20, 40, 80, 120, 240} {
		if i > 0 {
			goapp.Log.Infof("increase align size to %d", alignBuffer)
		}
		alignIDs, err = align(ocStrs, nStrs, alignBuffer)
		if err == nil {
			break
		}
	}
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
		for nTo == -1 && wTo < len(alignIDs) {
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

func align(oStrs []string, nStrs []string, step int) ([]int, error) {
	res := make([]int, len(oStrs))
	i, j := 0, 0
	addRes := func (partlyAl []int){
		for pi, v := range partlyAl {
			if v > -1 {
				res[i+pi] = partlyAl[pi] + j
			} else {
				res[i+pi] = -1
			}
		}
	}
	for i < len(oStrs) {
		partlyAl, partlyAlAlt := doPartlyAlign(oStrs[i:min(i+step, len(oStrs))], nStrs[j:min(j+step, len(nStrs))], step)
		addRes(partlyAl)
		if min(i+step, len(oStrs)) == len(oStrs) && min(j+step, len(nStrs)) == len(nStrs) {
			break
		}
		nextI, err := findLastConsequtiveAlign(partlyAl, oStrs[i:], nStrs[j:])
		if err != nil {
			goapp.Log.Infof("try alternative alignment")
			if partlyAlAlt != nil {
				partlyAl = partlyAlAlt
				nextI, err = findLastConsequtiveAlign(partlyAl, oStrs[i:], nStrs[j:])
			}
			if err != nil {
				return nil, errors.Wrapf(err, "can't find next steps starting at %d, %d", i, j)
			}
			addRes(partlyAl)
		}
		i = i + nextI
		j = j + partlyAl[nextI]
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

func doPartlyAlign(s1 []string, s2 []string, step int) ([]int, []int) {
	ind := func(i1, i2 int) int {
		return i1*step + i2
	}
	l1, l2 := len(s1), len(s2)
	h := make([]byte, step*step)
	hb := make([]moveType, step*step)

	// prn := func() {
	// 	sb := strings.Builder{}
	// 	for i := 0; i < step; i++ {
	// 		for j := 0; j < step; j++ {
	// 			if h[ind(i, j)] < 10 {
	// 				sb.WriteString("0")
	// 			}
	// 			sb.WriteString(fmt.Sprintf("%d ", h[ind(i, j)]))
	// 		}
	// 		sb.WriteString("\n")
	// 	}
	// 	fmt.Print(sb.String())
	// }
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
	// prn()
	prepareRes := func(i1From int) []int {
		res := make([]int, l1)
		i2 := l2 - 1
		for i1 := l1 - 1; i1 > i1From; i1-- {
			res[i1] = -1
		}
		for i1 := i1From; i1 >= 0; i1-- {
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

	// as this is partly alignment, it may not be an optimal aligment,
	// try to return the alternative alignment
	// if may work if many words are inserted into s2
	// h :=
	// 00 01 02 03 04 05 06 07 08 09
	// 01 01 02 03 04 05 06 07 08 09
	// 02 02 02 03 04 05 05 06 07 08
	// 03 03 03 03 04 05 06 06 07 08
	// 04 04 04 04 04 05 06 07 07 08
	// 05 05 05 05 05 05 06 07 08 08 <== use this as starting point
	// 06 06 06 06 06 06 06 07 08 09
	// 07 07 07 07 07 07 07 07 08 09 <== default starting point
	min := l1 - 1
	for i := l1 - 2; i > 0; i-- {
		if h[ind(i, l2-1)] < h[ind(min, l2-2)] {
			min = i
		}
	}
	if h[ind(min, l2-1)] < h[ind(l1-1, l2-1)] {
		return prepareRes(l1 - 1), prepareRes(min)
	}
	return prepareRes(l1 - 1), nil
}

// findLastConsequtiveAlign find the last two words matching in a row
func findLastConsequtiveAlign(partlyAl []int, oStrs []string, nStrs []string) (int, error) {

	for i := len(partlyAl) - 1; i > 1; i-- {
		v := partlyAl[i]
		if v > -1 && v-1 == partlyAl[i-1] {
			if oStrs[i-1] != nStrs[partlyAl[i-1]] {
				i--
			} else if oStrs[i] == nStrs[partlyAl[i]] {
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
