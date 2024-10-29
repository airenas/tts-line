package utils

import (
	"github.com/airenas/go-app/pkg/goapp"
	"github.com/pkg/errors"
)

func Align(s1 []string, s2 []string) ([]int, error) {
	var res []int
	var err error
	for i, alignBuffer := range [...]int{20, 40, 80, 120, 240} {
		if i > 0 {
			goapp.Log.Infof("increase align size to %d", alignBuffer)
		}
		res, err = align(s1, s2, alignBuffer)
		if err == nil {
			break
		}
	}
	return res, err
}

func align(oStrs []string, nStrs []string, step int) ([]int, error) {
	res := make([]int, len(oStrs))
	i, j := 0, 0
	addRes := func(partlyAl []int) {
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
	return -1, errors.New("no consecutive match found in partly aligned sequences")
}
