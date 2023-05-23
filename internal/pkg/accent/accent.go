package accent

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/pkg/errors"
)

// ToAccentString add accent to string
func ToAccentString(w string, a int) (string, error) {
	if a == 0 {
		return w, nil
	}
	pos := a%100 - 1
	tp := a / 100
	rn := []rune(w)
	if pos >= len(rn) || pos < 0 {
		return "", errors.Errorf("Wrong accent pos %d in %s", a, w)
	}
	as, err := toString(rn[pos], tp)
	if err != nil {
		return "", errors.Wrapf(err, "Wrong accent %s", w)
	}
	if pos == 0 {
		return as + string(rn[pos+1:]), nil
	}
	return string(rn[:pos]) + as + string(rn[pos+1:]), nil
}

// Value returns accent value as int or 0
func Value(r rune) int {
	if r == Kairinis {
		return 1
	}
	if r == Desininis {
		return 2
	}
	if r == Riestinis {
		return 3
	}
	return 0
}

const (
	//Kairinis accent type
	Kairinis = '\\'
	//Desininis accent type
	Desininis = '/'
	//Riestinis accent type
	Riestinis = '~'
)

// IsWordOrWithAccent checks if word is a correct one
func IsWordOrWithAccent(v string) bool {
	rns := []rune(v)
	for i := 0; i < len(rns); i++ {
		if i < (len(rns)-3) && rns[i] == '{' &&
			unicode.IsLetter(rns[i+1]) && Value(rns[i+2]) > 0 && rns[i+3] == '}' {
			i = i + 3
		} else if !unicode.IsLetter(rns[i]) {
			return false
		}
	}
	return len(rns) > 0
}

func ClearAccents(v string) string {
	rns := []rune(v)
	sb := strings.Builder{}
	for i := 0; i < len(rns); i++ {
		if i < (len(rns)-3) && rns[i] == '{' &&
			unicode.IsLetter(rns[i+1]) && Value(rns[i+2]) > 0 && rns[i+3] == '}' {
			sb.WriteRune(rns[i+1])
			i = i + 3
		} else {
			sb.WriteRune(rns[i])
		}
	}
	return sb.String()
}

func TranscriberAccent(acc int) string {
	if acc == 1 {
		return "4"
	}
	if acc == 2 {
		return "9"
	}
	if acc == 3 {
		return "3"
	}
	return ""
}

func toString(r rune, tp int) (string, error) {
	if tp == 1 {
		return fmt.Sprintf("{%c%c}", r, Kairinis), nil
	} else if tp == 2 {
		return fmt.Sprintf("{%c%c}", r, Desininis), nil
	} else if tp == 3 {
		return fmt.Sprintf("{%c%c}", r, Riestinis), nil
	}
	return "", errors.Errorf("Unknown accent type %d", tp)
}
