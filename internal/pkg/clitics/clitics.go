package clitics

import (
	"bufio"
	"io"
	"os"
	"strings"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/pkg/errors"
)

//ReadClitics from a file
func ReadClitics(fStr string) (map[string]bool, error) {
	file, err := os.Open(fStr)
	if err != nil {
		return nil, errors.Wrap(err, "unable to open file: "+fStr)
	}
	defer file.Close()
	return readClitics(file)
}

func readClitics(r io.Reader) (map[string]bool, error) {
	scanner := bufio.NewScanner(r)
	res := make(map[string]bool)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line != "" {
			res[line] = true
		}
	}
	if scanner.Err() != nil {
		return nil, errors.Wrap(scanner.Err(), "can read lines")
	}
	if len(res) == 0 {
		return nil, errors.New("no clitics loaded")
	}
	goapp.Log.Infof("Loaded %d clitics", len(res))
	return res, nil
}
