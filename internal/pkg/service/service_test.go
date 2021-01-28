package service

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"

	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/test/mocks"
	"github.com/airenas/tts-line/internal/pkg/test/mocks/matchers"
)

var (
	synthesizerMock *mocks.MockSynthesizer
	cnfMock         *mocks.MockConfigurator
)

func initTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	synthesizerMock = mocks.NewMockSynthesizer()
	cnfMock = mocks.NewMockConfigurator()
}

func TestWrongPath(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest("GET", "/invalid", nil)
	testCode(t, req, 404)
}

func TestWrongMethod(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest("GET", "/synthesize", nil)
	testCode(t, req, 405)
}

func Test_Returns(t *testing.T) {
	initTest(t)
	pegomock.When(cnfMock.Configure(matchers.AnyPtrToHttpRequest(), matchers.AnyPtrToApiInput())).
		ThenReturn(&api.TTSRequestConfig{Text: "olia1", OutputFormat: api.AudioMP3}, nil)
	pegomock.When(synthesizerMock.Work(matchers.AnyPtrToApiTTSRequestConfig())).
		ThenReturn(&api.Result{AudioAsString: "wav"}, nil)

	req := httptest.NewRequest("POST", "/synthesize", toReader(api.Input{Text: "olia"}))
	resp := testCode(t, req, 200)
	bytes, _ := ioutil.ReadAll(resp.Body)
	assert.Contains(t, string(bytes), `"audioAsString":"wav"`)
	txt := synthesizerMock.VerifyWasCalled(pegomock.Once()).Work(matchers.AnyPtrToApiTTSRequestConfig()).GetCapturedArguments()
	assert.Equal(t, "olia1", txt.Text)
	assert.Equal(t, "mp3", txt.OutputFormat.String())
	_, inp := cnfMock.VerifyWasCalled(pegomock.Once()).Configure(matchers.AnyPtrToHttpRequest(), matchers.AnyPtrToApiInput()).
		GetCapturedArguments()
	assert.Equal(t, "olia", inp.Text)
}

func Test_Fail(t *testing.T) {
	initTest(t)
	pegomock.When(cnfMock.Configure(matchers.AnyPtrToHttpRequest(), matchers.AnyPtrToApiInput())).
		ThenReturn(&api.TTSRequestConfig{Text: "olia1", OutputFormat: api.AudioMP3}, nil)
	pegomock.When(synthesizerMock.Work(matchers.AnyPtrToApiTTSRequestConfig())).ThenReturn(nil, errors.New("haha"))
	req := httptest.NewRequest("POST", "/synthesize", toReader(api.Input{Text: "olia"}))
	testCode(t, req, 500)
}

func Test_FailConfigure(t *testing.T) {
	initTest(t)
	pegomock.When(cnfMock.Configure(matchers.AnyPtrToHttpRequest(), matchers.AnyPtrToApiInput())).
		ThenReturn(nil, errors.New("No format mmp"))
	pegomock.When(synthesizerMock.Work(matchers.AnyPtrToApiTTSRequestConfig())).
		ThenReturn(&api.Result{AudioAsString: "wav"}, nil)

	req := httptest.NewRequest("POST", "/synthesize", toReader(api.Input{Text: "olia"}))
	resp := testCode(t, req, 400)
	assert.Equal(t, "No format mmp\n", resp.Body.String())
}

func Test_FailOnWrongInput(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest("POST", "/synthesize", strings.NewReader("text"))
	testCode(t, req, 400)
}

func TestCode(t *testing.T) {
	assert.Equal(t, 200, getCode(&api.Result{}))
	assert.Equal(t, 200, getCode(&api.Result{AudioAsString: "olia"}))
	assert.Equal(t, 400, getCode(&api.Result{ValidationFailures: []api.ValidateFailure{api.ValidateFailure{}}}))
}

func newTestRouter() *mux.Router {
	return NewRouter(newTestData())
}

func toReader(inData api.Input) io.Reader {
	bytes, _ := json.Marshal(inData)
	return strings.NewReader(string(bytes))
}

func newTestData() *Data {
	res := &Data{Processor: synthesizerMock, Configurator: cnfMock}
	return res
}

func testCode(t *testing.T, req *http.Request, code int) *httptest.ResponseRecorder {
	resp := httptest.NewRecorder()
	newTestRouter().ServeHTTP(resp, req)
	assert.Equal(t, code, resp.Code)
	return resp
}
