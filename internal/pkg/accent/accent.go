package accent

import (
	"github.com/pkg/errors"
)

//ToAccentString add accent to string
func ToAccentString(w string, a int) (string, error) {
	if a == 0 {
		return w, nil
	}
	pos := a % 100
	tp := a / 100
	rn := []rune(w)
	if pos > len(rn) || pos < 1 {
		return "", errors.Errorf("Wrong accent pos %d in %s", a, w)
	}
	as, err := toString(tp)
	if err != nil {
		return "", errors.Wrapf(err, "Wrong accent %s", w)
	}
	return string(rn[:pos]) + as + string(rn[pos:]), nil
}

const (
	//Kairinis accent type
	Kairinis = "{\\}"
	//Desininis accent type
	Desininis = "{/}"
	//Riestinis accent type
	Riestinis = "{~}"
)

func toString(tp int) (string, error) {
	if tp == 1 {
		return Kairinis, nil
	} else if tp == 2 {
		return Desininis, nil
	} else if tp == 3 {
		return Riestinis, nil
	}
	return "", errors.Errorf("Unknown accent type %d", tp)
}
