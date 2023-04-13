package service

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/airenas/tts-line/internal/pkg/clitics/service/api"
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
	wMock *mockClitWorker
)

func initTest(t *testing.T) {
	wMock = &mockClitWorker{}
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
	wMock.On("Process", mock.Anything).Return([]*api.CliticsOutput{}, nil)
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
	wMock.On("Process", mock.Anything).Return(nil, errors.New("err"))
	req := httptest.NewRequest(http.MethodPost, "/clitics", strings.NewReader(`[{"word":"aa"}]`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	tEcho.ServeHTTP(tRec, req)

	assert.Equal(t, http.StatusInternalServerError, tRec.Code)
}

func TestReturnsResponse(t *testing.T) {
	initTest(t)
	wMock.On("Process", mock.Anything).Return([]*api.CliticsOutput{{ID: 1, AccentType: "NONE"}}, nil)
	req := httptest.NewRequest(http.MethodPost, "/clitics", strings.NewReader(`[{"word":"aa"}]`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	tEcho.ServeHTTP(tRec, req)

	assert.Equal(t, http.StatusOK, tRec.Code)
	assert.Equal(t, "[{\"id\":1,\"accentType\":\"NONE\"}]", strings.TrimSpace(tRec.Body.String()))
}

type mockClitWorker struct{ mock.Mock }

func (m *mockClitWorker) Process(words []*api.CliticsInput) ([]*api.CliticsOutput, error) {
	args := m.Called(words)
	return mocks.To[[]*api.CliticsOutput](args.Get(0)), args.Error(1)
}
