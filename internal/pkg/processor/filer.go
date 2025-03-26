package processor

import (
	"context"
	"encoding/base64"
	"io"
	"os"
	"path"

	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type filer struct {
	dir   string
	fFile func(string) (io.WriteCloser, error)
}

// NewFiler creates new processor that save file for testing purposes
func NewFiler(dir string) (synthesizer.Processor, error) {
	res := &filer{}
	res.dir = dir
	res.fFile = func(name string) (io.WriteCloser, error) {
		f, err := os.Create(name)
		return f, err
	}
	return res, nil
}

func (p *filer) Process(ctx context.Context, data *synthesizer.TTSData) error {
	if data.Input.OutputFormat == api.AudioNone {
		return nil
	}
	return p.save(ctx, data.AudioMP3)
}

func (p *filer) save(ctx context.Context, data string) error {
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return err
	}

	fn := path.Join(p.dir, "out.mp3")
	log.Ctx(ctx).Debug().Msg("Saving " + fn)
	f, err := p.fFile(fn)
	if err != nil {
		return errors.Wrapf(err, "Can't open %s", fn)
	}
	defer f.Close()

	if _, err := f.Write(decoded); err != nil {
		return err
	}
	return nil
}
