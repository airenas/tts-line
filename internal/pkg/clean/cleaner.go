package clean

import (
	"html"
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

var multiSpaceRegexp *regexp.Regexp
var multiNewLineRegexp *regexp.Regexp

func init() {
	multiSpaceRegexp = regexp.MustCompile("[ ]{2,}")
	multiNewLineRegexp = regexp.MustCompile("[\n]{2,}")
}

//Text removes html tags from text
func Text(text string) string {
	return cleanHTML(text)
}

func cleanHTML(text string) string {
	p := bluemonday.StrictPolicy().AddSpaceWhenStrippingTag(true)
	t1 := strings.ReplaceAll(text, "<p>", "\n")
	t1 = strings.ReplaceAll(t1, "<br>", "\n")
	t1 = strings.ReplaceAll(t1, "</p>", "\n")

	res := p.Sanitize(t1)

	return fixSpacesNewLines(res)
}

func fixSpacesR(s string) string {
	return strings.TrimSpace(multiSpaceRegexp.ReplaceAllString(s, " "))
}

func fixSpacesNewLines(s string) string {
	sb := strings.Builder{}
	sp := ""
	for _, l := range strings.Split(strings.ReplaceAll(s, "\r", "\n"), "\n") {
		lf := fixSpacesR(changeSymbols(html.UnescapeString(l)))
		if lf != "" {
			sb.WriteString(sp)
			sb.WriteString(lf)
			sp = "\n"
		}
	}
	return sb.String()
}
