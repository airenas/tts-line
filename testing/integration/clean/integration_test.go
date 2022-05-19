//go:build integration
// +build integration

package clean

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/airenas/tts-line/testing/integration"
	"github.com/stretchr/testify/require"
)

type config struct {
	url        string
	httpclient *http.Client
}

var cfg config

func TestMain(m *testing.M) {
	cfg.url = integration.GetEnvOrFail("CLEAN_URL")
	cfg.httpclient = &http.Client{Timeout: time.Second * 20}

	tCtx, cf := context.WithTimeout(context.Background(), time.Second*20)
	defer cf()
	integration.WaitForOpenOrFail(tCtx, cfg.url)

	os.Exit(m.Run())
}

func TestLive(t *testing.T) {
	t.Parallel()
	integration.CheckCode(t,
		integration.Invoke(t, cfg.httpclient, integration.NewRequest(t, http.MethodGet, cfg.url, "/live", nil)), http.StatusOK)
}

type clData struct {
	Text string `json:"text"`
}

func Test_Success(t *testing.T) {
	t.Parallel()
	resp := integration.Invoke(t, cfg.httpclient, integration.NewRequest(t, http.MethodPost, cfg.url, "/clean",
		clData{Text: "Olia <html>"}))
	integration.CheckCode(t, resp, http.StatusOK)
	res := clData{}
	integration.Decode(t, resp, &res)
	require.Equal(t, "Olia", res.Text)
}

func Test_SpacesSuccess(t *testing.T) {
	t.Parallel()
	resp := integration.Invoke(t, cfg.httpclient, integration.NewRequest(t, http.MethodPost, cfg.url, "/clean",
		clData{Text: "Olia     <html>    ops\n\n\n\ntata"}))
	integration.CheckCode(t, resp, http.StatusOK)
	res := clData{}
	integration.Decode(t, resp, &res)
	require.Equal(t, "Olia ops\ntata", res.Text)
}
