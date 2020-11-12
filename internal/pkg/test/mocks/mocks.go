package mocks

import (
	"testing"

	"github.com/petergtz/pegomock"
)

//go:generate pegomock generate --package=mocks --output=synthesizer.go -m github.com/airenas/tts-line/internal/pkg/service Synthesizer

//go:generate pegomock generate --package=mocks --output=httpInvoker.go -m github.com/airenas/tts-line/internal/pkg/processor HTTPInvoker

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
