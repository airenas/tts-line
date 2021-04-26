package service

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/airenas/tts-line/internal/pkg/clitics/service/api"
	"github.com/airenas/tts-line/internal/pkg/test/mocks"
	"github.com/airenas/tts-line/internal/pkg/test/mocks/matchers"
	"github.com/labstack/echo/v4"
	"github.com/petergtz/pegomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var (
	tData *Data
	tEcho *echo.Echo
	tRec  *httptest.ResponseRecorder
	wMock *mocks.MockClitWorker
)

func initTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	wMock = mocks.NewMockClitWorker()
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
	pegomock.When(wMock.Process(matchers.AnySliceOfPtrToApiCliticsInput())).ThenReturn([]*api.CliticsOutput{}, nil)
	req := httptest.NewRequest(http.MethodPost, "/clitics", strings.NewReader(`[]`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	tEcho.ServeHTTP(tRec, req)

	assert.Equal(t, http.StatusOK, tRec.Code)
	assert.Equal(t, `[]`+"\n", tRec.Body.String())
}

func TestFails_Data(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest(http.MethodPost, "/clitics", strings.NewReader(`["word":"aa]`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	tEcho.ServeHTTP(tRec, req)

	assert.Equal(t, http.StatusBadRequest, tRec.Code)
}
func TestFails(t *testing.T) {
	initTest(t)
	pegomock.When(wMock.Process(matchers.AnySliceOfPtrToApiCliticsInput())).ThenReturn(nil, errors.New("err"))
	req := httptest.NewRequest(http.MethodPost, "/clitics", strings.NewReader(`[{"word":"aa"}]`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	tEcho.ServeHTTP(tRec, req)

	assert.Equal(t, http.StatusInternalServerError, tRec.Code)
}

func TestReturnsResponse(t *testing.T) {
	initTest(t)
	pegomock.When(wMock.Process(matchers.AnySliceOfPtrToApiCliticsInput())).ThenReturn([]*api.CliticsOutput{{ID: 1, AccentType: "NONE"}}, nil)
	req := httptest.NewRequest(http.MethodPost, "/clitics", strings.NewReader(`[{"word":"aa"}]`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	tEcho.ServeHTTP(tRec, req)

	assert.Equal(t, http.StatusOK, tRec.Code)
	assert.Equal(t, "[{\"id\":1,\"accentType\":\"NONE\"}]", strings.TrimSpace(tRec.Body.String()))
}
