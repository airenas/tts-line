package dtw

import (
	"context"
	"fmt"
	"math"

	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/rs/zerolog/log"
)

func Align(ctx context.Context, s1 []string, s2 []string) ([]int, error) {
	ctx, span := utils.StartSpan(ctx, "dtw.Align")
	defer span.End()

	var res []int
	var err error
	for i, band := range [...]int{40, 100, 200} {
		if i > 0 {
			log.Ctx(ctx).Debug().Msgf("increase align band to %d", band)
		}
		res, err = alignBanded(s1, s2, band)
		if err == nil {
			break
		}
	}
	return res, err
}

type moveType byte

const (
	start moveType = iota
	left
	top
	corner
	eqLeft
	eqTop
	eqStart
	eqCorner
	eqCorner2InARow
	eqCorner3InARow
)

func (m moveType) add(eq int) moveType {
	if eq != 0 {
		return m
	}
	switch m {
	case start:
		return eqStart
	case left:
		return eqLeft
	case top:
		return eqTop
	default:
		return m
	}
}

func (m moveType) addCorner(eq int) moveType {
	if eq != 0 {
		return corner
	}
	switch m {
	case corner:
		return eqCorner
	case eqCorner:
		return eqCorner2InARow
	case eqCorner2InARow:
		return eqCorner3InARow
	case eqCorner3InARow:
		return eqCorner3InARow
	case eqTop:
		return eqCorner2InARow
	case eqLeft:
		return eqCorner2InARow
	default:
		return eqCorner
	}
}

type rowMeta struct {
	jMin int
}

type dpValues struct {
	m        int
	rowsMeta []rowMeta
	band     int
	backs    []moveType
}

func (d *dpValues) calcNextValue(i int, j int, eq int, prevRow, currRow []int) (int, moveType) {
	if i == 0 && j == 0 {
		return eq, start.add(eq)
	}
	prevJ := j + 1 // ignore on first row
	if i > 0 {
		prevJ = d.rowsMeta[i-1].jMin
	}
	currJ := d.rowsMeta[i].jMin
	lv := add(atOrMax(currRow, j-currJ-1), 1)
	tv := atOrMax(prevRow, j-prevJ)
	if j != 0 || eq != 0 {
		tv = add(tv, 1)
	}
	cv := add(atOrMax(prevRow, j-prevJ-1), eq)
	if lv <= tv && lv <= cv {
		return lv, left.add(eq)
	}
	if cv <= tv {
		return cv, d.backAt(i-1, j-1).addCorner(eq)
	}
	return tv, top.add(eq)
}

func (d *dpValues) backAt(i, j int) moveType {
	if i < 0 || j < d.rowsMeta[i].jMin {
		return start
	}
	return d.backs[d.index(i, j-d.rowsMeta[i].jMin)]
}

func add(i1, i2 int) int {
	if i1 == math.MaxInt || i2 == math.MaxInt {
		return math.MaxInt
	}
	return i1 + i2
}

func atOrMax(currRow []int, i int) int {
	if i < 0 || i >= len(currRow) {
		return math.MaxInt
	}
	return currRow[i]
}

func (d *dpValues) index(i int, at int) int {
	return i*d.band + at
}

func alignBanded[T comparable](orig, pred []T, bandWidth int) ([]int, error) {
	n := len(orig)
	m := len(pred)
	res := make([]int, n)
	if n == 0 || m == 0 {
		for i := range res {
			res[i] = -1
		}
		return res, nil
	}
	dp := &dpValues{
		m:        m,
		rowsMeta: make([]rowMeta, n),
		band:     minInt(bandWidth, m),
		backs:    make([]moveType, n*minInt(bandWidth, m)),
	}
	prevRow := make([]int, dp.band)
	currRow := make([]int, dp.band)

	for i := 0; i < n; i++ {
		dp.rowsMeta[i].jMin = getFromIndex(dp, i, prevRow)
		rMeta := dp.rowsMeta[i]
		for j := rMeta.jMin; j < rMeta.jMin+dp.band; j++ {
			ind := j - rMeta.jMin
			eq := 1
			if orig[i] == pred[j] {
				eq = 0
			}
			currRow[ind], dp.backs[dp.index(i, ind)] = dp.calcNextValue(i, j, eq, prevRow, currRow)
		}
		copy(prevRow, currRow)
	}

	// Backtrack to recover alignment
	if dp.band+dp.rowsMeta[n-1].jMin < m {
		return nil, fmt.Errorf("alignment failed due to band width limit: %d < %d", dp.band+dp.rowsMeta[n-1].jMin, m)
	}

	bestJ := m - 1
	for i := n - 1; i >= 0; i-- {
		rMeta := dp.rowsMeta[i]
		res[i] = bestJ
		if i == 0 && bestJ < 0 {
			break
		}
		move := dp.backs[dp.index(i, bestJ-rMeta.jMin)]
		switch move {
		case corner:
			bestJ--
		case eqCorner:
			bestJ--
		case eqCorner2InARow:
			bestJ--
		case eqCorner3InARow:
			bestJ--
		case left:
			for (move == left || move == eqLeft) && bestJ > 0 {
				bestJ--
				res[i] = bestJ
				move = dp.backs[dp.index(i, bestJ-rMeta.jMin)]
			}
		case eqLeft:
			for (move == left || move == eqLeft) && bestJ > 0 {
				bestJ--
				move = dp.backs[dp.index(i, bestJ-rMeta.jMin)]
			}
		case top:
			res[i] = -1
		case eqTop: // do nothing
			if bestJ == 0 {
				bestJ--
			}
		case start:
			res[i] = 0
		}
	}
	return res, nil
}

func getFromIndex(dp *dpValues, i int, prevRow []int) int {
	if i == 0 {
		return 0
	}
	if dp.band >= dp.m {
		return 0
	}
	jMin := dp.rowsMeta[i-1].jMin
	for ind := jMin; ind < jMin+dp.band; ind++ {
		if dp.backAt(i-1, ind) == eqCorner3InARow {
			return minInt(dp.m-dp.band, ind)
		}
	}

	min := minAt(prevRow) + dp.rowsMeta[i-1].jMin - dp.band/4
	if min < 0 {
		return 0
	}
	return minInt(dp.m-dp.band, min)
}

// if equal some then return the last min
func minAt(prevRow []int) int {
	minVal, minI := math.MaxInt, 0
	for i, v := range prevRow {
		if i == 0 || v <= minVal {
			minVal = v
			minI = i
		}
	}
	return minI
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
