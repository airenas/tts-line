package wrapservice

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/airenas/tts-line/internal/pkg/wrapservice/api"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/airenas/go-app/pkg/goapp"
)

type (
	//WaveSynthesizer main sythesis processor
	WaveSynthesizer interface {
		Work(ctx context.Context, params *api.Params) (*api.Result, error)
	}
	//Data is service operation data
	Data struct {
		Port          int
		Processor     WaveSynthesizer
		HealthHandler http.Handler
	}
)

// StartWebServer starts the HTTP service and listens for synthesize requests
func StartWebServer(data *Data) error {
	goapp.Log.Info().Msgf("Starting HTTP TTS AM-VOC Wrapper Service at %d", data.Port)
	portStr := strconv.Itoa(data.Port)

	e := initRoutes(data)

	e.Server.Addr = ":" + portStr
	e.Server.IdleTimeout = time.Minute * 3
	e.Server.ReadHeaderTimeout = 15 * time.Second
	e.Server.ReadTimeout = 60 * time.Second
	e.Server.WriteTimeout = 180 * time.Second

	gracehttp.SetLogger(log.New(goapp.Log, "", 0))

	return gracehttp.Serve(e.Server)
}

var p *prometheus.Prometheus

func initRoutes(data *Data) *echo.Echo {
	e := echo.New()
	if p == nil {
		p = prometheus.NewPrometheus("avw", nil)
	}
	p.Use(e)

	e.POST("/synthesize", handleSynthesize(data))
	e.GET("/live", live(data))

	goapp.Log.Info().Msg("Routes:")
	for _, r := range e.Routes() {
		goapp.Log.Info().Msgf("  %s %s", r.Method, r.Path)
	}
	return e
}

func handleSynthesize(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		defer goapp.Estimate("Service method: synthesize")()

		ctype := c.Request().Header.Get(echo.HeaderContentType)
		if !strings.HasPrefix(ctype, echo.MIMEApplicationJSON) {
			goapp.Log.Error().Msg("Wrong content type")
			return echo.NewHTTPError(http.StatusBadRequest, "Wrong content type. Expected '"+echo.MIMEApplicationJSON+"'")
		}
		var input api.Input
		if err := c.Bind(&input); err != nil {
			goapp.Log.Error().Err(err).Send()
			return echo.NewHTTPError(http.StatusBadRequest, "Can read data")
		}
		if input.Text == "" {
			goapp.Log.Error().Msg("No text")
			return echo.NewHTTPError(http.StatusBadRequest, "No text")
		}
		if input.Voice == "" {
			goapp.Log.Error().Msg("No voice")
			return echo.NewHTTPError(http.StatusBadRequest, "No voice")
		}

		res, err := data.Processor.Work(ctx, &api.Params{Text: input.Text, Speed: input.Speed, Voice: input.Voice, Priority: input.Priority})
		if err != nil {
			goapp.Log.Error().Err(errors.Wrap(err, "cannot process text")).Send()
			return echo.NewHTTPError(http.StatusInternalServerError, "Cannot process text")
		}
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		c.Response().WriteHeader(http.StatusOK)
		return json.NewEncoder(c.Response()).Encode(res)
	}
}

func live(data *Data) func(echo.Context) error {
	return echo.WrapHandler(data.HealthHandler)
}
