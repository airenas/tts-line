package processor

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/pkg/errors"
)

type tagger struct {
	httpclient *http.Client
	url        string
}

//NewTagger creates new processor
func NewTagger(urlStr string) (synthesizer.Processor, error) {
	res := &tagger{}
	var err error
	urlRes, err := url.Parse(urlStr)
	if err != nil {
		return nil, errors.Wrap(err, "Can't parse url "+urlStr)
	}
	if urlRes.Host == "" {
		return nil, errors.New("Can't parse url " + urlStr)
	}
	res.url = urlRes.String()
	res.httpclient = &http.Client{}
	return res, nil
}

func (p *tagger) Process(data *synthesizer.TTSData) error {
	goapp.Log.Debugf("In: '%s'", data.TextWithNumbers)
	tagged, err := p.callTag(data.Text)
	if err != nil {
		return err
	}
	data.Words = mapTagResult(tagged)
	//goapp.Log.Debugf("Out: '%s'", data.TextWithNumbers)
	return nil
}

//TaggedWord - tagger's result
type TaggedWord struct {
	Type   string
	String string
	Mi     string
	Lemma  string
}

//Get return duration by calling the service
func (p *tagger) callTag(text string) ([]*TaggedWord, error) {
	req, err := http.NewRequest("POST", p.url, strings.NewReader(text))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "text/plain")

	goapp.Log.Debugf("Sending text to: %s", p.url)
	resp, err := p.httpclient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, errors.New("Can't tag")
	}
	var res []*TaggedWord
	err = decodeJSONAndLog(resp.Body, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func mapTagResult(tags []*TaggedWord) []*synthesizer.ProcessedWord {
	res := make([]*synthesizer.ProcessedWord, 0)
	for _, t := range tags {
		if t.Type != "SPACE" {
			pw := synthesizer.ProcessedWord{Tagged: mapTag(t)}
			res = append(res, &pw)
		}
	}
	return res
}

func mapTag(tag *TaggedWord) synthesizer.TaggedWord {
	res := synthesizer.TaggedWord{}
	if tag.Type == "SEPARATOR" {
		res.Separator = tag.String
	} else if tag.Type == "WORD" {
		res.Word = tag.String
		res.Lemma = tag.Lemma
		res.Mi = tag.Mi
	}
	return res
}
