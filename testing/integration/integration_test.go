//go:build integration
// +build integration

package integration

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type config struct {
	url          string
	semantikaURL string
	httpclient   *http.Client
}

var cfg config

func TestMain(m *testing.M) {
	cfg.url = GetEnvOrFail("TTS_URL")
	cfg.semantikaURL = GetEnvOrFail("MORPHOLOGY_URL")
	cfg.httpclient = &http.Client{Timeout: time.Second * 20}

	tCtx, cf := context.WithTimeout(context.Background(), time.Second*20)
	defer cf()
	WaitForOpenOrFail(tCtx, cfg.url)
	WaitForOpenOrFail(tCtx, cfg.semantikaURL)

	//start mocks service for private services - not in this docker compose
	l, ts := startMockService(9876)
	defer ts.Close()
	defer l.Close()

	os.Exit(m.Run())
}

func TestLive(t *testing.T) {
	t.Parallel()
	CheckCode(t, Invoke(t, cfg.httpclient, NewRequest(t, http.MethodGet, cfg.url, "/live", nil)), http.StatusOK)
}

func TestSynthesize_FailVoice(t *testing.T) {
	t.Parallel()
	resp := Invoke(t, cfg.httpclient, NewRequest(t, http.MethodPost, cfg.url, "/synthesize",
		api.Input{Text: "Olia", Voice: "xxx"}))
	CheckCode(t, resp, http.StatusBadRequest)
}

func TestSynthesize_OK_SaveRequest(t *testing.T) {
	t.Parallel()
	resp := Invoke(t, cfg.httpclient, NewRequest(t, http.MethodPost, cfg.url, "/synthesize",
		api.Input{Text: "Olia", Voice: "astra", AllowCollectData: &[]bool{true}[0]}))
	CheckCode(t, resp, http.StatusOK)
}

func TestSynthesize_Fail_SaveRequest(t *testing.T) {
	t.Parallel()
	req := NewRequest(t, http.MethodPost, cfg.url, "/synthesize", api.Input{Text: "Olia", Voice: "astra", AllowCollectData: &[]bool{false}[0]})
	req.Header.Add("x-tts-collect-data", "always")
	resp := Invoke(t, cfg.httpclient, req)
	CheckCode(t, resp, http.StatusBadRequest)
}

func TestSynthesize_Success(t *testing.T) {
	t.Parallel()
	resp := Invoke(t, cfg.httpclient, NewRequest(t, http.MethodPost, cfg.url, "/synthesize",
		api.Input{Text: "Olia xrytas", Voice: "astra"}))
	CheckCode(t, resp, http.StatusOK)
	res := api.Result{}
	Decode(t, resp, &res)
	assert.Empty(t, res.Audio)
	require.NotEmpty(t, res.AudioAsString)
}

func TestSynthesize_MsgPack_Success(t *testing.T) {
	t.Parallel()
	req := NewRequest(t, http.MethodPost, cfg.url, "/synthesize",
		api.Input{Text: "Olia xrytas", Voice: "astra"})
	req.Header.Add(echo.HeaderAccept, echo.MIMEApplicationMsgpack)
	resp := Invoke(t, cfg.httpclient, req)
	CheckCode(t, resp, http.StatusOK)
	res := api.Result{}
	DecodeMsgPack(t, resp, &res)
	assert.Empty(t, res.AudioAsString)
	require.NotEmpty(t, res.Audio)
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
	testSSML(t, `<speak><p/><voice name="astra">Olia</voice><p/><voice name="laimis">Olia</voice></speak>`,
		http.StatusOK)
}

func TestSynthesize_SSMLOK_Word(t *testing.T) {
	t.Parallel()
	testSSML(t, `<speak><p/><intelektika:w acc="{O/}lia">Olia</intelektika:w></speak>`,
		http.StatusOK)
}

func TestSynthesize_SSMLFail_WrongAcc(t *testing.T) {
	t.Parallel()
	testSSML(t, `<speak><p/><intelektika:w acc="{)-}lia">Olia</intelektika:w></speak>`,
		http.StatusBadRequest)
}

func TestSynthesizeCustom_Success(t *testing.T) {
	t.Parallel()
	resp := Invoke(t, cfg.httpclient, NewRequest(t, http.MethodPost, cfg.url, "/synthesize",
		api.Input{Text: "Olia", Voice: "astra", AllowCollectData: &[]bool{true}[0], OutputTextFormat: "accented"}))
	CheckCode(t, resp, http.StatusOK)
	res := api.Result{}
	Decode(t, resp, &res)
	require.NotEmpty(t, res.AudioAsString)
	require.NotEmpty(t, res.RequestID)

	resp = Invoke(t, cfg.httpclient, NewRequest(t, http.MethodPost, cfg.url,
		fmt.Sprintf("/synthesizeCustom?requestID=%s", res.RequestID),
		api.Input{Text: "Olia", Voice: "astra"}))
	CheckCode(t, resp, http.StatusOK)
	res = api.Result{}
	Decode(t, resp, &res)
	require.NotEmpty(t, res.AudioAsString)
}

func TestSynthesizeCustom_Fail_Differs(t *testing.T) {
	t.Parallel()
	resp := Invoke(t, cfg.httpclient, NewRequest(t, http.MethodPost, cfg.url, "/synthesize",
		api.Input{Text: "Olia", Voice: "astra", AllowCollectData: &[]bool{true}[0], OutputTextFormat: "accented"}))
	CheckCode(t, resp, http.StatusOK)
	res := api.Result{}
	Decode(t, resp, &res)
	require.NotEmpty(t, res.AudioAsString)
	require.NotEmpty(t, res.RequestID)

	resp = Invoke(t, cfg.httpclient, NewRequest(t, http.MethodPost, cfg.url,
		fmt.Sprintf("/synthesizeCustom?requestID=%s", res.RequestID),
		api.Input{Text: "Olia olia olia", Voice: "astra"}))
	CheckCode(t, resp, http.StatusBadRequest)
	res = api.Result{}
	Decode(t, resp, &res)
}

func TestSynthesizeCustom_FailNoID(t *testing.T) {
	t.Parallel()
	resp := Invoke(t, cfg.httpclient, NewRequest(t, http.MethodPost, cfg.url,
		fmt.Sprintf("/synthesizeCustom?requestID=%s", "xxx"),
		api.Input{Text: "Olia", Voice: "astra"}))
	CheckCode(t, resp, http.StatusBadRequest)
}

func TestSynthesizeCustom_FailSSML(t *testing.T) {
	t.Parallel()
	resp := Invoke(t, cfg.httpclient, NewRequest(t, http.MethodPost, cfg.url, "/synthesize",
		api.Input{Text: "Olia", Voice: "astra", AllowCollectData: &[]bool{true}[0], OutputTextFormat: "accented"}))
	CheckCode(t, resp, http.StatusOK)
	res := api.Result{}
	Decode(t, resp, &res)
	require.NotEmpty(t, res.AudioAsString)
	require.NotEmpty(t, res.RequestID)

	resp = Invoke(t, cfg.httpclient, NewRequest(t, http.MethodPost, cfg.url,
		fmt.Sprintf("/synthesizeCustom?requestID=%s", res.RequestID),
		api.Input{Text: "<speak>aaa</speak>", Voice: "astra"}))
	CheckCode(t, resp, http.StatusBadRequest)
}

func TestRequest_Success(t *testing.T) {
	t.Parallel()
	resp := Invoke(t, cfg.httpclient, NewRequest(t, http.MethodPost, cfg.url, "/synthesize",
		api.Input{Text: "Olia", Voice: "astra", AllowCollectData: &[]bool{true}[0], OutputTextFormat: "accented"}))
	CheckCode(t, resp, http.StatusOK)
	res := api.Result{}
	Decode(t, resp, &res)
	require.NotEmpty(t, res.AudioAsString)
	require.NotEmpty(t, res.RequestID)

	for i := int64(0); i < 10; i++ {
		resp = Invoke(t, cfg.httpclient, NewRequest(t, http.MethodGet, cfg.url,
			fmt.Sprintf("/request/%s", res.RequestID), nil))
		CheckCode(t, resp, http.StatusOK)
		resI := api.InfoResult{}
		Decode(t, resp, &resI)
		require.Equal(t, i, resI.Count, "count is not increased")

		resp = Invoke(t, cfg.httpclient, NewRequest(t, http.MethodPost, cfg.url,
			fmt.Sprintf("/synthesizeCustom?requestID=%s", res.RequestID),
			api.Input{Text: "Olia", Voice: "astra"}))
		CheckCode(t, resp, http.StatusOK)
	}
}

func testSSML(t *testing.T, in string, exp int) {
	t.Helper()
	resp := Invoke(t, cfg.httpclient, NewRequest(t, http.MethodPost, cfg.url, "/synthesize",
		api.Input{Text: in, Voice: "astra", TextType: "ssml"}))
	CheckCode(t, resp, exp)
	if exp == http.StatusOK {
		res := api.Result{}
		Decode(t, resp, &res)
		require.NotEmpty(t, res.AudioAsString)
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
		case "/mock-am":
			b, err := os.ReadFile("data/test.wav")
			if err != nil {
				log.Print(err.Error())
			}
			var input struct {
				Text            string    `json:"text"`
				DurationsChange []float64 `json:"durationsChange"`
			}
			if r.Body != nil {
				defer r.Body.Close()
				if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
					log.Printf("failed to parse input: %v", err)
				} else {
					log.Printf("Received input: text=%s", input.Text)
				}
			}
			res_durations := make([]int, len(input.DurationsChange) + 1)
			for i := range res_durations {
				res_durations[i] = 1
			}
			respObj := map[string]interface{}{
				"data":      base64.StdEncoding.EncodeToString(b),
				"step":      256,
				"durations": res_durations,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(respObj)
		}
	}))

	ts.Listener.Close()
	ts.Listener = l

	// Start the server.
	ts.Start()
	log.Printf("started mock srv on port: %d", port)
	return l, ts
}
