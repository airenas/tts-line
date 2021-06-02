package utils

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHTTPCreate(t *testing.T) {
	_, err := NewHTTWrap("")
	assert.NotNil(t, err)
	hw, err := NewHTTWrap("http://local:8080")
	assert.NotNil(t, hw)
	assert.Nil(t, err)
}

type testType struct {
	Test string `json:"test"`
}

func TestInvokeText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "POST", req.Method)
		br, _ := ioutil.ReadAll(req.Body)
		assert.Equal(t, "olia", string(br))
		rw.Write([]byte(`{"test":"respo"}`))
	}))
	defer server.Close()
	hw, _ := NewHTTWrap(server.URL)
	lg := ""
	hw.flog = func(st, data string) {
		lg = lg + st + data
	}
	var tt testType
	err := hw.InvokeText("olia", &tt)
	assert.Nil(t, err)
	assert.Equal(t, "respo", tt.Test)
	assert.Equal(t, "Input : oliaCall : "+server.URL+"Output: {\"test\":\"respo\"}", lg)
}

func TestInvokeJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "POST", req.Method)
		br, _ := ioutil.ReadAll(req.Body)
		assert.Equal(t, "{\"test\":\"haha\"}\n", string(br))
		rw.Write([]byte(`{"test":"respo"}`))
	}))
	defer server.Close()
	hw, _ := NewHTTWrap(server.URL)
	lg := ""
	hw.flog = func(st, data string) {
		lg = lg + st + data
	}
	var tt testType
	err := hw.InvokeJSON(testType{Test: "haha"}, &tt)
	assert.Nil(t, err)
	assert.Equal(t, "respo", tt.Test)
	assert.Equal(t, "Input : {\"test\":\"haha\"}\nCall : "+server.URL+"Output: {\"test\":\"respo\"}", lg)
}

func TestInvokeFail_Server(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(400)
	}))
	defer server.Close()
	hw, _ := NewHTTWrap(server.URL)
	var tt testType
	err := hw.InvokeText("olia", &tt)
	assert.NotNil(t, err)
}

func TestInvokeFail_Response(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte(`{"test":"respo"`))
	}))
	defer server.Close()
	hw, _ := NewHTTWrap(server.URL)
	var tt testType
	err := hw.InvokeText("olia", &tt)
	assert.NotNil(t, err)
}

func TestTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte(`{"test":"respo"}`))
	}))
	defer server.Close()
	hw, _ := NewHTTWrap(server.URL)
	hw.Timeout = time.Millisecond * 50
	var tt testType
	err := hw.InvokeText("olia", &tt)
	assert.Nil(t, err)
	err = hw.InvokeJSON(testType{Test: "haha"}, &tt)
	assert.Nil(t, err)
}

func TestTimeout_Fail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte(`{"test":"respo"}`))
		time.Sleep(time.Second)
	}))
	defer server.Close()
	hw, _ := NewHTTWrap(server.URL)
	hw.Timeout = time.Millisecond * 50
	var tt testType
	err := hw.InvokeText("olia", &tt)
	assert.NotNil(t, err)
	err = hw.InvokeJSON(testType{Test: "haha"}, &tt)
	assert.NotNil(t, err)
}
