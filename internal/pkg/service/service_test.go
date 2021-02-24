package service

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/petergtz/pegomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/test/mocks"
	"github.com/airenas/tts-line/internal/pkg/test/mocks/matchers"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/labstack/echo/v4"
)

var (
	synthesizerMock *mocks.MockSynthesizer
	cnfMock         *mocks.MockConfigurator
)

var (
	tData *Data
	tEcho *echo.Echo
	tResp *httptest.ResponseRecorder
)

func initTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	synthesizerMock = mocks.NewMockSynthesizer()
	cnfMock = mocks.NewMockConfigurator()
	tData = newTestData()
	tEcho = initRoutes(tData)
	tResp = httptest.NewRecorder()
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
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
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
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 500)
}

func Test_FailConfigure(t *testing.T) {
	initTest(t)
	pegomock.When(cnfMock.Configure(matchers.AnyPtrToHttpRequest(), matchers.AnyPtrToApiInput())).
		ThenReturn(nil, errors.New("No format mmp"))
	pegomock.When(synthesizerMock.Work(matchers.AnyPtrToApiTTSRequestConfig())).
		ThenReturn(&api.Result{AudioAsString: "wav"}, nil)

	req := httptest.NewRequest("POST", "/synthesize", toReader(api.Input{Text: "olia"}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	resp := testCode(t, req, 400)
	assert.Equal(t, `{"message":"No format mmp"}`+"\n", resp.Body.String())
}

func Test_FailOnWrongInput(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest("POST", "/synthesize", strings.NewReader("text"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 400)
}

func TestCustom_Returns(t *testing.T) {
	initTest(t)
	pegomock.When(cnfMock.Configure(matchers.AnyPtrToHttpRequest(), matchers.AnyPtrToApiInput())).
		ThenReturn(&api.TTSRequestConfig{Text: "olia1", OutputFormat: api.AudioMP3}, nil)
	pegomock.When(synthesizerMock.Work(matchers.AnyPtrToApiTTSRequestConfig())).
		ThenReturn(&api.Result{AudioAsString: "wav"}, nil)

	req := httptest.NewRequest("POST", "/synthesizeCustom?requestID=1", toReader(api.Input{Text: "olia"}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
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

func TestCustom_SetAllowCollectData(t *testing.T) {
	initTest(t)
	pegomock.When(cnfMock.Configure(matchers.AnyPtrToHttpRequest(), matchers.AnyPtrToApiInput())).
		ThenReturn(&api.TTSRequestConfig{Text: "olia1", OutputFormat: api.AudioMP3}, nil)
	pegomock.When(synthesizerMock.Work(matchers.AnyPtrToApiTTSRequestConfig())).
		ThenReturn(&api.Result{AudioAsString: "wav"}, nil)
	req := httptest.NewRequest("POST", "/synthesizeCustom?requestID=1", toReader(api.Input{Text: "olia"}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 200)
	inp := synthesizerMock.VerifyWasCalled(pegomock.Once()).Work(matchers.AnyPtrToApiTTSRequestConfig()).
		GetCapturedArguments()
	assert.Equal(t, true, inp.AllowCollectData)
}

func TestCustom_FailWithWrongAllowCollectData(t *testing.T) {
	initTest(t)
	pegomock.When(cnfMock.Configure(matchers.AnyPtrToHttpRequest(), matchers.AnyPtrToApiInput())).
		ThenReturn(&api.TTSRequestConfig{Text: "olia1", OutputFormat: api.AudioMP3}, nil)
	pegomock.When(synthesizerMock.Work(matchers.AnyPtrToApiTTSRequestConfig())).
		ThenReturn(&api.Result{AudioAsString: "wav"}, nil)
	b := false
	req := httptest.NewRequest("POST", "/synthesizeCustom?requestID=1",
		toReader(api.Input{Text: "olia", AllowCollectData: &b}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 400)
}

func TestCustom_AcceptsAllowCollectData(t *testing.T) {
	initTest(t)
	pegomock.When(cnfMock.Configure(matchers.AnyPtrToHttpRequest(), matchers.AnyPtrToApiInput())).
		ThenReturn(&api.TTSRequestConfig{Text: "olia1", OutputFormat: api.AudioMP3}, nil)
	pegomock.When(synthesizerMock.Work(matchers.AnyPtrToApiTTSRequestConfig())).
		ThenReturn(&api.Result{AudioAsString: "wav"}, nil)
	b := true
	req := httptest.NewRequest("POST", "/synthesizeCustom?requestID=1",
		toReader(api.Input{Text: "olia", AllowCollectData: &b}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 200)
}

func TestCustom_Fail(t *testing.T) {
	initTest(t)
	pegomock.When(cnfMock.Configure(matchers.AnyPtrToHttpRequest(), matchers.AnyPtrToApiInput())).
		ThenReturn(&api.TTSRequestConfig{Text: "olia1", OutputFormat: api.AudioMP3}, nil)
	pegomock.When(synthesizerMock.Work(matchers.AnyPtrToApiTTSRequestConfig())).ThenReturn(nil, errors.New("haha"))
	req := httptest.NewRequest("POST", "/synthesizeCustom?requestID=1", toReader(api.Input{Text: "olia"}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 500)
}

func TestCustom_Fail400(t *testing.T) {
	initTest(t)
	pegomock.When(cnfMock.Configure(matchers.AnyPtrToHttpRequest(), matchers.AnyPtrToApiInput())).
		ThenReturn(&api.TTSRequestConfig{Text: "olia1", OutputFormat: api.AudioMP3}, nil)
	pegomock.When(synthesizerMock.Work(matchers.AnyPtrToApiTTSRequestConfig())).ThenReturn(nil, utils.ErrNoRecord)
	req := httptest.NewRequest("POST", "/synthesizeCustom?requestID=1", toReader(api.Input{Text: "olia"}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 400)
}

func TestCustom_FailConfigure(t *testing.T) {
	initTest(t)
	pegomock.When(cnfMock.Configure(matchers.AnyPtrToHttpRequest(), matchers.AnyPtrToApiInput())).
		ThenReturn(nil, errors.New("No format mmp"))
	pegomock.When(synthesizerMock.Work(matchers.AnyPtrToApiTTSRequestConfig())).
		ThenReturn(&api.Result{AudioAsString: "wav"}, nil)

	req := httptest.NewRequest("POST", "/synthesizeCustom?requestID=1", toReader(api.Input{Text: "olia"}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	resp := testCode(t, req, 400)
	assert.Equal(t, `{"message":"No format mmp"}`+"\n", resp.Body.String())
}

func TestCustom_FailOnWrongInput(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest("POST", "/synthesizeCustom?requestID=1", strings.NewReader("text"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 400)
}

func TestCustom_FailNoRequest(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest("POST", "/synthesizeCustom", toReader(api.Input{Text: "olia"}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 400)
}

func TestCode(t *testing.T) {
	assert.Equal(t, 200, getCode(&api.Result{}))
	assert.Equal(t, 200, getCode(&api.Result{AudioAsString: "olia"}))
	assert.Equal(t, 400, getCode(&api.Result{ValidationFailures: []api.ValidateFailure{{}}}))
}

func TestBadReqError(t *testing.T) {
	tests := []struct {
		v  error
		e  bool
		es string
	}{
		{v: errors.New("olia"), e: false, es: ""},
		{v: utils.ErrNoRecord, e: true, es: "RequestID not found"},
		{v: utils.ErrTextDoesNotMatch, e: true, es: "Original text does not match the modified"},
		{v: utils.NewErrBadAccent([]string{"olia"}), e: true, es: "Bad accents: [olia]"},
		{v: errors.Wrap(utils.NewErrBadAccent([]string{"olia"}), "test"), e: true, es: "Bad accents: [olia]"},
	}

	for i, tc := range tests {
		v, str := badReqError(tc.v)
		assert.Equal(t, tc.e, v, "Fail %d", i)
		assert.Equal(t, tc.es, str, "Fail %d", i)
	}
}

func toReader(inData api.Input) io.Reader {
	bytes, _ := json.Marshal(inData)
	return strings.NewReader(string(bytes))
}

func newTestData() *Data {
	res := &Data{SyntData: PrData{Processor: synthesizerMock, Configurator: cnfMock},
		SyntCustomData: PrData{Processor: synthesizerMock, Configurator: cnfMock}}
	return res
}

func testCode(t *testing.T, req *http.Request, code int) *httptest.ResponseRecorder {
	tEcho.ServeHTTP(tResp, req)
	assert.Equal(t, code, tResp.Code)
	return tResp
}
