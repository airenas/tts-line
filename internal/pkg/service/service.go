package service

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/facebookgo/grace/gracehttp"

	"github.com/airenas/go-app/pkg/goapp"

	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type (
	//Configurator configures request from header and request and configuration
	Configurator interface {
		Configure(*http.Request, *api.Input) (*api.TTSRequestConfig, error)
	}

	//Synthesizer main sythesis processor
	Synthesizer interface {
		Work(*api.TTSRequestConfig) (*api.Result, error)
	}
	//Data is service operation data
	Data struct {
		Port         int
		Processor    Synthesizer
		Configurator Configurator
	}
)

//StartWebServer starts the HTTP service and listens for the admin requests
func StartWebServer(data *Data) error {
	goapp.Log.Infof("Starting HTTP TTS Line service at %d", data.Port)
	portStr := strconv.Itoa(data.Port)

	e := initRoutes(data)

	e.Server.Addr = ":" + portStr

	w := goapp.Log.Writer()
	defer w.Close()
	l := log.New(w, "", 0)
	gracehttp.SetLogger(l)

	return gracehttp.Serve(e.Server)
}

func initRoutes(data *Data) *echo.Echo {
	e := echo.New()
	e.Use(middleware.Logger())
	p := prometheus.NewPrometheus("tts", nil)
	p.Use(e)

	e.POST("/synthesize", synthesizeText(data))
	e.GET("/live", live(data))

	goapp.Log.Info("Routes:")
	for _, r := range e.Routes() {
		goapp.Log.Infof("  %s %s", r.Method, r.Path)
	}
	return e
}

func synthesizeText(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		defer goapp.Estimate("Service synthesize method")()

		ctype := c.Request().Header.Get(echo.HeaderContentType)
		if !strings.HasPrefix(ctype, echo.MIMEApplicationJSON) {
			goapp.Log.Error("Wrong content type")
			return echo.NewHTTPError(http.StatusBadRequest, "Wrong content type. Expected '"+echo.MIMEApplicationJSON+"'")
		}
		inp := new(api.Input)
		if err := c.Bind(inp); err != nil {
			goapp.Log.Error(err)
			return echo.NewHTTPError(http.StatusBadRequest, "Cannot decode input")
		}

		cfg, err := data.Configurator.Configure(c.Request(), inp)
		if err != nil {
			goapp.Log.Error("Cannot prepare request config " + err.Error())
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		resp, err := data.Processor.Work(cfg)
		if err != nil {
			goapp.Log.Error("Can't process. ", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Service error")
		}

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		c.Response().WriteHeader(getCode(resp))
		enc := json.NewEncoder(c.Response())
		enc.SetEscapeHTML(false)
		return enc.Encode(resp)
	}
}

func getCode(resp *api.Result) int {
	if len(resp.ValidationFailures) > 0 {
		return http.StatusBadRequest
	}
	return 200
}

func live(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		return c.JSONBlob(http.StatusOK, []byte(`{"service":"OK"}`))
	}
}
