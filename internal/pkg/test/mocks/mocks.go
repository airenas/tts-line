package mocks

import (
	aapi "github.com/airenas/tts-line/internal/pkg/acronyms/service/api"
	tapi "github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/stretchr/testify/mock"
)

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

type Synthesizer struct{ mock.Mock }

func (m *Synthesizer) Work(in *tapi.TTSRequestConfig) (*tapi.Result, error) {
	args := m.Called(in)
	return To[*tapi.Result](args.Get(0)), args.Error(1)
}

type HTTPInvokerJSON struct{ mock.Mock }

func (m *HTTPInvokerJSON) InvokeText(in string, out interface{}) error {
	args := m.Called(in, out)
	return args.Error(0)
}

func (m *HTTPInvokerJSON) InvokeJSON(in interface{}, out interface{}) error {
	args := m.Called(in, out)
	return args.Error(0)
}

func (m *HTTPInvokerJSON) InvokeJSONU(URL string, in interface{}, out interface{}) error {
	args := m.Called(URL, in, out)
	return args.Error(0)
}
