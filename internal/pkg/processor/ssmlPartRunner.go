package processor

import (
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
)

type SSMLPartRunner struct {
	processors []synthesizer.Processor
}

//NewJoinAudio joins results into one audio
func NewSSMLPartRunner(processors []synthesizer.Processor) *SSMLPartRunner {
	return &SSMLPartRunner{processors: processors}
}

func (p *SSMLPartRunner) Process(data *synthesizer.TTSData) error {
	for _, part := range data.SSMLParts {
		switch part.Cfg.Type {
		case synthesizer.SSMLText:
			for _, pr := range p.processors {
				if err := pr.Process(part); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
