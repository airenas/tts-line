package processor

import (
	"fmt"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
)

// SSMLPartRunner runs all processors for Text part
type SSMLPartRunner struct {
	processors []synthesizer.Processor
}

// NewSSMLPartRunner creates runner for SSML parts representing Text
func NewSSMLPartRunner(processors []synthesizer.Processor) *SSMLPartRunner {
	return &SSMLPartRunner{processors: processors}
}

// Process main method
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

// Info return info about processor
func (p *SSMLPartRunner) Info() string {
	return fmt.Sprintf("SSMLPartRunner(%d)", len(p.processors))
}
