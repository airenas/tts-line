package wrapservice

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"

	"github.com/airenas/tts-line/internal/pkg/test/mocks"
	"github.com/airenas/tts-line/internal/pkg/wrapservice/api"
)

var (
	synthesizerMock *mocks.MockWaveSynthesizer
	tData           *Data
	tEcho           *echo.Echo
	tRec            *httptest.ResponseRecorder
)

func initTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	synthesizerMock = mocks.NewMockWaveSynthesizer()
	var hf http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }
	tData = &Data{Port: 8000, Processor: synthesizerMock, HealthHandler: hf}
	tEcho = initRoutes(tData)
	tRec = httptest.NewRecorder()
}

func TestWrongPath(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest("GET", "/invalid", nil)
	testCode(t, req, 404)
}

func TestLive(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest("GET", "/live", nil)
	testCode(t, req, 200)
}

func TestLive503(t *testing.T) {
	initTest(t)
	var hf http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusServiceUnavailable) }
	tData = &Data{Port: 8000, Processor: synthesizerMock, HealthHandler: hf}
	tEcho = initRoutes(tData)
	tRec = httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/live", nil)
	testCode(t, req, 503)
}

func TestMatrics(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest("GET", "/metrics", nil)
	testCode(t, req, 200)
}

func TestWrongMethod(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest("GET", "/synthesize", nil)
	testCode(t, req, 405)
}

func Test_Returns(t *testing.T) {
	initTest(t)
	pegomock.When(synthesizerMock.Work(pegomock.AnyString(), pegomock.AnyFloat32(), pegomock.AnyString())).ThenReturn("wav", nil)
	req := httptest.NewRequest("POST", "/synthesize", toReader(api.Input{Text: "olia", Speed: 0.9, Voice: "aa"}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	resp := testCode(t, req, 200)
	bytes, _ := ioutil.ReadAll(resp.Body)
	assert.Contains(t, string(bytes), `"data":"wav"`)
	txt, sp, voice := synthesizerMock.VerifyWasCalled(pegomock.Once()).Work(pegomock.AnyString(), pegomock.AnyFloat32(), pegomock.AnyString()).GetCapturedArguments()
	assert.Equal(t, "olia", txt)
	assert.InDelta(t, 0.9, sp, 0.0001)
	assert.Equal(t, "aa", voice)
}

func Test_Fail(t *testing.T) {
	initTest(t)
	pegomock.When(synthesizerMock.Work(pegomock.AnyString(), pegomock.AnyFloat32(), pegomock.AnyString())).ThenReturn("", errors.New("haha"))
	req := httptest.NewRequest("POST", "/synthesize", toReader(api.Input{Text: "olia", Voice: "aa"}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 500)
}

func Test_FailOnWrongInput(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest("POST", "/synthesize", strings.NewReader("text"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 400)
	req = httptest.NewRequest("POST", "/synthesize", toReader(api.Input{Text: "olia", Voice: ""}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 400)
	req = httptest.NewRequest("POST", "/synthesize", toReader(api.Input{Text: "", Voice: "aaa"}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 400)

}

func toReader(inData api.Input) io.Reader {
	bytes, _ := json.Marshal(inData)
	return strings.NewReader(string(bytes))
}

func testCode(t *testing.T, req *http.Request, code int) *httptest.ResponseRecorder {
	tRec = httptest.NewRecorder()
	tEcho.ServeHTTP(tRec, req)
	assert.Equal(t, code, tRec.Code)
	return tRec
}
