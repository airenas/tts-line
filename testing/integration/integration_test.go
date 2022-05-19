//go:build integration
// +build integration

package integration

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

type config struct {
	url          string
	semantikaURL string
	httpclient   *http.Client
}

var cfg config

func TestMain(m *testing.M) {
	cfg.url = getEnvOrFail("TTS_URL")
	cfg.semantikaURL = getEnvOrFail("MORPHOLOGY_URL")
	cfg.httpclient = &http.Client{Timeout: time.Second * 20}

	tCtx, cf := context.WithTimeout(context.Background(), time.Second*20)
	defer cf()
	waitForOpenOrFail(tCtx, cfg.url)
	waitForOpenOrFail(tCtx, cfg.semantikaURL)

	//start mocks service for private services - not in this docker compose
	l, ts := startMockService(9876)
	defer ts.Close()
	defer l.Close()

	os.Exit(m.Run())
}

func waitForOpenOrFail(ctx context.Context, URL string) {
	u, err := url.Parse(URL)
	if err != nil {
		log.Fatalf("FAIL: can't parse %s", URL)
	}
	for {
		err = listen(net.JoinHostPort(u.Hostname(), u.Port()))
		if err == nil {
			return
		}
		select {
		case <-ctx.Done():
			log.Fatalf("FAIL: can't access %s", URL)
			break
		case <-time.After(500 * time.Millisecond):
		}
	}
}

func getEnvOrFail(s string) string {
	res := os.Getenv(s)
	if res == "" {
		log.Fatalf("no env '%s'", s)
	}
	return res
}

func listen(urlStr string) error {
	log.Printf("dial %s", urlStr)
	conn, err := net.DialTimeout("tcp", urlStr, time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()
	return nil
}

func TestLive(t *testing.T) {
	t.Parallel()
	checkCode(t, invoke(t, newRequest(t, http.MethodGet, cfg.url, "/live", nil)), http.StatusOK)
}

func TestSynthesize_FailVoice(t *testing.T) {
	t.Parallel()
	resp := invoke(t, newRequest(t, http.MethodPost, cfg.url, "/synthesize",
		api.Input{Text: "Olia", Voice: "xxx"}))
	checkCode(t, resp, http.StatusBadRequest)
}

func TestSynthesize_Success(t *testing.T) {
	t.Parallel()
	resp := invoke(t, newRequest(t, http.MethodPost, cfg.url, "/synthesize",
		api.Input{Text: "Olia", Voice: "astra"}))
	checkCode(t, resp, http.StatusOK)
	res := api.Result{}
	decode(t, resp, &res)
	require.NotEmpty(t, res.AudioAsString)
}

func TestSynthesize_SSMLFail_NoSSML(t *testing.T) {
	t.Parallel()
	testSSML(t, "olia", http.StatusBadRequest)
}

func TestSynthesize_SSMLFail_WrongSSML(t *testing.T) {
	t.Parallel()
	testSSML(t, "<speak>Olia</speak>olia", http.StatusBadRequest)
}

func TestSynthesize_SSMLFail_NoText(t *testing.T) {
	t.Parallel()
	testSSML(t, `<speak><p/><voice name="astra"></voice></speak>`, http.StatusBadRequest)
}

func TestSynthesize_SSMLOK_Voice(t *testing.T) {
	t.Parallel()
	testSSML(t, `<speak><p/><voice name="astra">Olia</voice></speak>`, http.StatusOK)
}

func TestSynthesize_SSMLOK_Several(t *testing.T) {
	t.Parallel()
	testSSML(t, `<speak><p/><voice name="astra">Olia</voice><p/><voice name="linas">Olia</voice></speak>`,
		http.StatusOK)
}

func TestSynthesizeCustom_Success(t *testing.T) {
	t.Parallel()
	resp := invoke(t, newRequest(t, http.MethodPost, cfg.url, "/synthesize",
		api.Input{Text: "Olia", Voice: "astra", AllowCollectData: &[]bool{true}[0], OutputTextFormat: "accented"}))
	checkCode(t, resp, http.StatusOK)
	res := api.Result{}
	decode(t, resp, &res)
	require.NotEmpty(t, res.AudioAsString)
	require.NotEmpty(t, res.RequestID)

	resp = invoke(t, newRequest(t, http.MethodPost, cfg.url,
		fmt.Sprintf("/synthesizeCustom?requestID=%s", res.RequestID),
		api.Input{Text: "Olia", Voice: "astra"}))
	checkCode(t, resp, http.StatusOK)
	res = api.Result{}
	decode(t, resp, &res)
	require.NotEmpty(t, res.AudioAsString)
}

func TestSynthesizeCustom_FailNoID(t *testing.T) {
	t.Parallel()
	resp := invoke(t, newRequest(t, http.MethodPost, cfg.url,
		fmt.Sprintf("/synthesizeCustom?requestID=%s", "xxx"),
		api.Input{Text: "Olia", Voice: "astra"}))
	checkCode(t, resp, http.StatusBadRequest)
}

func TestRequest_Success(t *testing.T) {
	t.Parallel()
	resp := invoke(t, newRequest(t, http.MethodPost, cfg.url, "/synthesize",
		api.Input{Text: "Olia", Voice: "astra", AllowCollectData: &[]bool{true}[0], OutputTextFormat: "accented"}))
	checkCode(t, resp, http.StatusOK)
	res := api.Result{}
	decode(t, resp, &res)
	require.NotEmpty(t, res.AudioAsString)
	require.NotEmpty(t, res.RequestID)

	for i := int64(0); i < 10; i++ {
		resp = invoke(t, newRequest(t, http.MethodGet, cfg.url,
			fmt.Sprintf("/request/%s", res.RequestID), nil))
		checkCode(t, resp, http.StatusOK)
		resI := api.InfoResult{}
		decode(t, resp, &resI)
		require.Equal(t, i, resI.Count, "count is not increased") 
		
		resp = invoke(t, newRequest(t, http.MethodPost, cfg.url,
			fmt.Sprintf("/synthesizeCustom?requestID=%s", res.RequestID),
			api.Input{Text: "Olia", Voice: "astra"}))
		checkCode(t, resp, http.StatusOK)
	}
}

func testSSML(t *testing.T, in string, exp int) {
	t.Helper()
	resp := invoke(t, newRequest(t, http.MethodPost, cfg.url, "/synthesize",
		api.Input{Text: in, Voice: "astra", TextType: "ssml"}))
	checkCode(t, resp, exp)
	if exp == http.StatusOK {
		res := api.Result{}
		decode(t, resp, &res)
		require.NotEmpty(t, res.AudioAsString)
	}
}

func newRequest(t *testing.T, method string, srv, urlSuffix string, body interface{}) *http.Request {
	t.Helper()
	req, err := http.NewRequest(method, srv+urlSuffix, toReader(body))
	require.Nil(t, err, "not nil error = %v", err)
	if body != nil {
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
	}
	return req
}

func toReader(data interface{}) io.Reader {
	bytes, _ := json.Marshal(data)
	return strings.NewReader(string(bytes))
}

func invoke(t *testing.T, r *http.Request) *http.Response {
	t.Helper()
	resp, err := cfg.httpclient.Do(r)
	require.Nil(t, err, "not nil error = %v", err)
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func checkCode(t *testing.T, resp *http.Response, expected int) {
	t.Helper()
	if resp.StatusCode != expected {
		b, _ := ioutil.ReadAll(resp.Body)
		require.Equal(t, expected, resp.StatusCode, string(b))
	}
}

func startMockService(port int) (net.Listener, *httptest.Server) {
	// create a listener with the desired port.
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("can't start mock service: %v", err)
	}
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// log.Printf("request to: " + r.URL.String())
		switch r.URL.String() {
		case "/mock-number-replace":
			io.Copy(w, strings.NewReader(`"Olia"`))
		case "/mock-obscene-filter":
			io.Copy(w, strings.NewReader(`[{"token":"Olia","obscene":0}]`))
		case "/mock-compare":
			io.Copy(w, strings.NewReader(`{"rc":1,"badacc":[]}`))	
		case "/mock-am":
			b, err := ioutil.ReadFile("data/test.wav")
			if err != nil {
				log.Printf(err.Error())
			}
			io.Copy(w, strings.NewReader(fmt.Sprintf(`{"data":"%s"}`, base64.StdEncoding.EncodeToString(b))))
		}
	}))

	ts.Listener.Close()
	ts.Listener = l

	// Start the server.
	ts.Start()
	log.Printf("started mock srv on port: %d", port)
	return l, ts
}

func decode(t *testing.T, resp *http.Response, to interface{}) {
	t.Helper()
	require.Nil(t, json.NewDecoder(resp.Body).Decode(to))
}
