package clean

import (
	"strings"

	emoji "github.com/tmdvs/Go-Emoji-Utils"
)

func dropEmojis(s string) string {
	matches := emoji.FindAll(s)

	er := make(map[rune]bool)
	for _, item := range matches {
		emo := item.Match.(emoji.Emoji)
		for _, r := range []rune(emo.Value) {
			er[r] = true
		}
	}

	res := strings.Builder{}
	lr := '-'
	for _, r := range s {
		if er[r] {
			if lr != ' ' {
				lr = ' '
				res.WriteRune(lr)
			}
			continue
		} 
		res.WriteRune(r)
		lr = r
	}
	return res.String()
}
