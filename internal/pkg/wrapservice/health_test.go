package wrapservice

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHealthHandler(t *testing.T) {
	h, err := NewHealthHandler("localhost", "localhost")
	assert.Nil(t, err)
	assert.NotNil(t, h)
}

func TestNewHealthHandler_Fail(t *testing.T) {
	h, err := NewHealthHandler("", "localhost")
	assert.NotNil(t, err)
	assert.Nil(t, h)
	h, err = NewHealthHandler("localhost", "")
	assert.NotNil(t, err)
	assert.Nil(t, h)
}

func Test200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	h, _ := NewHealthHandler(server.URL, server.URL)
	tRec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/live", nil)
	h.ServeHTTP(tRec, req)
	assert.Equal(t, http.StatusOK, tRec.Code)
	assert.Contains(t, tRec.Body.String(), "AM")
	assert.Contains(t, tRec.Body.String(), "Vocoder")
}

func Test503(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()
	server2 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}))
	defer server2.Close()

	test503(t, server.URL, server.URL)
	test503(t, server2.URL, server.URL)
	test503(t, server.URL, server2.URL)
}

func test503(t *testing.T, url1, url2 string) {
	h, _ := NewHealthHandler(url1, url2)
	tRec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/live", nil)
	h.ServeHTTP(tRec, req)
	assert.Equal(t, 503, tRec.Code)
}
