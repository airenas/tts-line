package processor

import (
	"encoding/base64"
	"os"
	"path"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
)

type filer struct {
	dir string
}

//NewFiler creates new processor
func NewFiler(dir string) (synthesizer.Processor, error) {
	res := &filer{}
	res.dir = dir
	return res, nil
}

func (p *filer) Process(data *synthesizer.TTSData) error {
	err := p.save(data.AudioMP3)
	if err != nil {
		return err
	}
	return nil
}

func (p *filer) save(data string) error {
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return err
	}

	f, err := os.Create(path.Join(p.dir, "out.mp3"))
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write(decoded); err != nil {
		return err
	}
	return nil
}
