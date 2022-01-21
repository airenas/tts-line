package clitics

import (
	"bufio"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/pkg/errors"
)

type phrase struct {
	word    string
	isLemma bool
	accent  int
}

//Phrases processor
type Phrases struct {
	wordMap map[string][][]*phrase
}

//ReadPhrases from a file
func ReadPhrases(fStr string) (*Phrases, error) {
	file, err := os.Open(fStr)
	if err != nil {
		return nil, errors.Wrap(err, "unable to open file: "+fStr)
	}
	defer file.Close()
	return readPhrases(file)
}

func readPhrases(r io.Reader) (*Phrases, error) {
	res := &Phrases{wordMap: make(map[string][][]*phrase)}
	scanner := bufio.NewScanner(r)
	pc := 0
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "//") {
			phr, err := readLine(line)
			if err != nil {
				return nil, errors.Wrapf(err, "can't read line '%s'", line)
			}
			res.wordMap[phr[0].word] = append(res.wordMap[phr[0].word], phr)
			pc++
		}
	}
	if scanner.Err() != nil {
		return nil, errors.Wrap(scanner.Err(), "can read lines")
	}
	for k, v := range res.wordMap {
		sort.Slice(v, func(i, j int) bool { return len(v[i]) > len(v[j]) })
		res.wordMap[k] = v
	}
	goapp.Log.Infof("Loaded %d phrases", pc)
	return res, nil
}

func readLine(l string) ([]*phrase, error) {
	res := make([]*phrase, 0)
	words := strings.Split(strings.SplitN(l, ",", 2)[0], " ")
	for _, w := range words {
		if w != "" {
			ph := &phrase{}
			var err error
			ph.word, ph.isLemma, ph.accent, err = parse(w)
			if err != nil {
				return nil, errors.Wrapf(err, "can't parse word '%s'", w)
			}
			res = append(res, ph)
		}
	}
	if len(res) < 2 {
		return nil, errors.New("too short phrase")
	}
	return res, nil
}

func parse(w string) (word string, lemma bool, accent int, err error) {
	word, accent, err = getAccent(w)
	if strings.HasSuffix(w, ":l") {
		word = word[:len(word)-2]
		lemma = true
	}
	return word, lemma, accent, err
}

func getAccent(w string) (word string, accent int, err error) {
	for i, a := range "493" {
		in := strings.Index(w, string(a))
		if in > -1 {
			return w[:in] + w[in+1:], (i+1)*100 + in, nil
		}
	}
	return w, 0, nil
}
