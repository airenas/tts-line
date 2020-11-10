package synthesizer

import (
	"errors"

	"github.com/airenas/tts-line/internal/pkg/service/api"
)

//MainWorker does synthesis work
type MainWorker struct {
}

//Work is main method
func (mw *MainWorker) Work(text string) (*api.Result, int, error) {
	return nil, 0, errors.New("Not implemented")
}
