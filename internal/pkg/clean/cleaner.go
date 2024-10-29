package clean

import (
	"regexp"
	"strings"

	"github.com/airenas/go-app/pkg/goapp"
	pclean "github.com/airenas/tts-line/pkg/clean"
	"github.com/microcosm-cc/bluemonday"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
)

var multiSpaceRegexp *regexp.Regexp

// var multiNewLineRegexp *regexp.Regexp

func init() {
	multiSpaceRegexp = regexp.MustCompile("[ ]{2,}")
	// multiNewLineRegexp = regexp.MustCompile("[\n]{2,}")
}

// Text removes html tags from text
func Text(text string) string {
	return cleanHTML(text)
}

func cleanHTML(text string) string {
	nText, err := skipIgnoreTags(text)
	if err != nil {
		goapp.Log.Error().Err(err).Send()
		nText = text
	}
	p := bluemonday.StrictPolicy().AddSpaceWhenStrippingTag(true)
	t1 := strings.ReplaceAll(nText, "<p>", "\n")
	t1 = strings.ReplaceAll(t1, "<br>", "\n")
	t1 = strings.ReplaceAll(t1, "</p>", "\n")

	res := p.Sanitize(t1)

	return fixSpacesNewLines(res)
}

func skipIgnoreTags(text string) (string, error) {
	doc, err := html.Parse(strings.NewReader(text))
	if err != nil {
		return "", errors.Wrapf(err, "Can't parse text as html doc")
	}
	checkRemoveChildNodes(doc)
	return docToString(doc), nil
}

func docToString(doc *html.Node) string {
	res := strings.Builder{}
	_ = html.Render(&res, doc)
	return res.String()
}

func checkRemoveChildNodes(n *html.Node) {
	c := n.FirstChild
	for c != nil {
		nc := c.NextSibling
		if skipNode(c) {
			n.RemoveChild(c)
		}
		c = nc
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		checkRemoveChildNodes(c)
	}
}

func skipNode(n *html.Node) bool {
	if n.Type == html.ElementNode {
		for _, a := range n.Attr {
			if a.Key == "speak" && a.Val == "ignore" {
				return true
			}
		}
	}
	return false
}

func fixSpacesR(s string) string {
	return strings.TrimSpace(multiSpaceRegexp.ReplaceAllString(s, " "))
}

func fixSpacesNewLines(s string) string {
	sb := strings.Builder{}
	sp := ""
	for _, l := range strings.Split(strings.ReplaceAll(s, "\r", "\n"), "\n") {
		lf := fixSpacesR(pclean.ChangeSymbols(pclean.DropEmojis(html.UnescapeString(l))))
		if lf != "" {
			sb.WriteString(sp)
			sb.WriteString(lf)
			sp = "\n"
		}
	}
	return sb.String()
}
