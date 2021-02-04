package service

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

var (
	tData *Data
	tEcho *echo.Echo
	tReq  *http.Request
	tRec  *httptest.ResponseRecorder
)

func initTest(t *testing.T) {
	tData = &Data{Port: 8000}
	tEcho = initRoutes(tData)
	tReq = httptest.NewRequest(http.MethodPost, "/clean", strings.NewReader(`{"text":"Tekstas"}`))
	tReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	tRec = httptest.NewRecorder()
}

func TestLive(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest(http.MethodGet, "/live", nil)

	e := initRoutes(tData)
	e.ServeHTTP(tRec, req)
	assert.Equal(t, http.StatusOK, tRec.Code)
	assert.Equal(t, `{"service":"OK"}`, tRec.Body.String())
}

func TestClean(t *testing.T) {
	initTest(t)

	tEcho.ServeHTTP(tRec, tReq)

	assert.Equal(t, http.StatusOK, tRec.Code)
	assert.Equal(t, `{"text":"Tekstas"}`+"\n", tRec.Body.String())
}

func TestClean_WorksCleaning(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest(http.MethodPost, "/clean", strings.NewReader(`{"text":"<html>tekstas<p> <em>kita eilutė</html>"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	tEcho.ServeHTTP(tRec, req)

	assert.Equal(t, http.StatusOK, tRec.Code)
	assert.Equal(t, `{"text":"tekstas\nkita eilutė"}`+"\n", tRec.Body.String())
}

func TestClean_WorksCleaningLessThan(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest(http.MethodPost, "/clean", strings.NewReader(`{"text":"olia &lt;"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	tEcho.ServeHTTP(tRec, req)

	assert.Equal(t, http.StatusOK, tRec.Code)
	assert.Equal(t, `{"text":"olia <"}`+"\n", tRec.Body.String())
}

func TestConvert_FailData(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest(http.MethodPost, "/clean", strings.NewReader(`{"text":"wrong data}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	tEcho.ServeHTTP(tRec, req)

	assert.Equal(t, http.StatusBadRequest, tRec.Code)
}

func TestConvert_FailWrongType(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest(http.MethodPost, "/clean", strings.NewReader(`{"text":"ok"}`))
	tEcho.ServeHTTP(tRec, req)

	assert.Equal(t, http.StatusBadRequest, tRec.Code)
}
