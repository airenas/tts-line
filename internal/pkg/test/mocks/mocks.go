package mocks

import (
	"testing"

	"github.com/petergtz/pegomock"
	"github.com/stretchr/testify/mock"
	aapi "github.com/airenas/tts-line/internal/pkg/acronyms/service/api"
)

//go:generate pegomock generate --package=mocks --output=synthesizer.go -m github.com/airenas/tts-line/internal/pkg/service Synthesizer

//go:generate pegomock generate --package=mocks --output=infoGetter.go -m github.com/airenas/tts-line/internal/pkg/service InfoGetter

//go:generate pegomock generate --package=mocks --output=configurator.go -m github.com/airenas/tts-line/internal/pkg/service Configurator

//go:generate pegomock generate --package=mocks --output=httpInvoker.go -m github.com/airenas/tts-line/internal/pkg/processor HTTPInvoker

//go:generate pegomock generate --package=mocks --output=httpInvokerJSOM.go -m github.com/airenas/tts-line/internal/pkg/processor HTTPInvokerJSON

//go:generate pegomock generate --package=mocks --output=waveSynthesizer.go -m github.com/airenas/tts-line/internal/pkg/wrapservice WaveSynthesizer

//go:generate pegomock generate --package=mocks --output=saverDB.go -m github.com/airenas/tts-line/internal/pkg/processor SaverDB

//go:generate pegomock generate --package=mocks --output=loadDB.go -m github.com/airenas/tts-line/internal/pkg/processor LoadDB

//go:generate pegomock generate --package=mocks --output=audioLoader.go -m github.com/airenas/tts-line/internal/pkg/processor AudioLoader

//go:generate pegomock generate --package=mocks --output=exporter.go -m github.com/airenas/tts-line/internal/pkg/exporter Exporter

//AttachMockToTest register pegomock verification to be passed to testing engine
func AttachMockToTest(t *testing.T) {
	pegomock.RegisterMockFailHandler(handleByTest(t))
}

func handleByTest(t *testing.T) pegomock.FailHandler {
	return func(message string, callerSkip ...int) {
		if message != "" {
			t.Error(message)
		}
	}
}

// To convert interface to object
func To[T interface{}](val interface{}) T {
	if val == nil {
		var res T
		return res
	}
	return val.(T)
}

type Worker struct{ mock.Mock }

func (m *Worker) Process(word, mi string) ([]aapi.ResultWord, error) {
	args := m.Called(word, mi)
	return To[[]aapi.ResultWord](args.Get(0)), args.Error(1)
}
