package service

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/airenas/tts-line/internal/pkg/acronyms/service/api"
	"github.com/airenas/tts-line/internal/pkg/test/mocks"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	tData *Data
	tEcho *echo.Echo
	tRec  *httptest.ResponseRecorder
	wMock *mocks.Worker
)

func initTest(t *testing.T) {
	wMock = &mocks.Worker{}
	tData = &Data{Port: 8000, Worker: wMock}
	tEcho = initRoutes(tData)
	tRec = httptest.NewRecorder()
}

func TestLive(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest(http.MethodGet, "/live", nil)

	tEcho.ServeHTTP(tRec, req)
	assert.Equal(t, http.StatusOK, tRec.Code)
	assert.Equal(t, `{"service":"OK"}`, tRec.Body.String())
}

func TestNotFound(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest(http.MethodPost, "/any", strings.NewReader(``))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	tEcho.ServeHTTP(tRec, req)

	assert.Equal(t, http.StatusNotFound, tRec.Code)
}

func TestProvides(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest(http.MethodPost, "/acronyms", strings.NewReader(`[]`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	tEcho.ServeHTTP(tRec, req)

	assert.Equal(t, http.StatusOK, tRec.Code)
	assert.Equal(t, `[]`+"\n", tRec.Body.String())
}

func TestFails_Data(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest(http.MethodPost, "/acronyms", strings.NewReader(`["word":"aa]`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	tEcho.ServeHTTP(tRec, req)

	assert.Equal(t, http.StatusBadRequest, tRec.Code)
}
func TestFails(t *testing.T) {
	initTest(t)
	wMock.On("Process", mock.Anything, mock.Anything).Return(nil, errors.New("err"))
	req := httptest.NewRequest(http.MethodPost, "/acronyms", strings.NewReader(`[{"word":"aa"}]`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	tEcho.ServeHTTP(tRec, req)

	assert.Equal(t, http.StatusInternalServerError, tRec.Code)
}

func TestReturnsResponse(t *testing.T) {
	initTest(t)
	wMock.On("Process", mock.Anything, mock.Anything).Return([]api.ResultWord{{Word: "aa"}}, nil)
	req := httptest.NewRequest(http.MethodPost, "/acronyms", strings.NewReader(`[{"word":"aa"}]`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	tEcho.ServeHTTP(tRec, req)

	assert.Equal(t, http.StatusOK, tRec.Code)
	assert.Equal(t, "[{\"words\":[{\"word\":\"aa\"}]}]", strings.TrimSpace(tRec.Body.String()))
}

func TestGETFailEmpty(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest(http.MethodGet, "/acronym/", nil)

	tEcho.ServeHTTP(tRec, req)

	assert.Equal(t, http.StatusNotFound, tRec.Code)
}

func TestGETFails(t *testing.T) {
	initTest(t)
	wMock.On("Process", mock.Anything, mock.Anything).Return(nil, errors.New("err"))
	req := httptest.NewRequest(http.MethodGet, "/acronym/aa", nil)

	tEcho.ServeHTTP(tRec, req)

	assert.Equal(t, http.StatusInternalServerError, tRec.Code)
}

func TestGETReturnsResponse(t *testing.T) {
	initTest(t)
	wMock.On("Process", mock.Anything, mock.Anything).Return([]api.ResultWord{{Word: "aa"}}, nil)
	req := httptest.NewRequest(http.MethodGet, "/acronym/aa", nil)

	tEcho.ServeHTTP(tRec, req)

	assert.Equal(t, http.StatusOK, tRec.Code)
	assert.Equal(t, "[{\"word\":\"aa\"}]", strings.TrimSpace(tRec.Body.String()))
}
