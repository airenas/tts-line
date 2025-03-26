package processor

import (
	"context"
	"fmt"
	"strings"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
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
func (p *SSMLPartRunner) Process(ctx context.Context, data *synthesizer.TTSData) error {
	ctx, span := utils.StartSpan(ctx, "SSMLPartRunner.Process")
	defer span.End()

	for _, part := range data.SSMLParts {
		switch part.Cfg.Type {
		case synthesizer.SSMLText:
			for _, pr := range p.processors {
				if err := pr.Process(ctx, part); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Info return info about processor
func (p *SSMLPartRunner) Info() string {
	prInfo := strings.Builder{}
	for _, pr := range p.processors {
		s := utils.RetrieveInfo(pr)
		if s != "" {
			prInfo.WriteString(s)
			prInfo.WriteString("\n")
		}
	}
	return fmt.Sprintf("SSMLPartRunner(%d)[\n%s\n]", len(p.processors), strings.TrimSpace(prInfo.String()))
}
