package service

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/clean"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
)

type (
	//Data is service operation data
	Data struct {
		Port int
	}
)

//StartWebServer starts the HTTP service and listens for the convert requests
func StartWebServer(data *Data) error {
	goapp.Log.Infof("Starting HTTP audio convert service at %d", data.Port)
	portStr := strconv.Itoa(data.Port)

	e := initRoutes(data)

	if err := e.Start(":" + portStr); err != nil {
		return errors.Wrap(err, "Can't start HTTP listener at port "+portStr)
	}
	return nil
}

func initRoutes(data *Data) *echo.Echo {
	e := echo.New()
	e.Use(middleware.Logger())
	p := prometheus.NewPrometheus("echo", nil)
	p.Use(e)

	e.POST("/clean", handleClean(data))
	e.GET("/live", live(data))

	goapp.Log.Info("Routes:")
	for _, r := range e.Routes() {
		goapp.Log.Infof("  %s %s", r.Method, r.Path)
	}
	return e
}

type input struct {
	Text string `json:"text"`
}

type output struct {
	Text string `json:"text"`
}

func handleClean(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		defer goapp.Estimate("Service clean method")()

		ctype := c.Request().Header.Get(echo.HeaderContentType)
		if !strings.HasPrefix(ctype, echo.MIMEApplicationJSON) {
			goapp.Log.Error("Wrong content type")
			return echo.NewHTTPError(http.StatusBadRequest, "Wrong content type. Expected '"+echo.MIMEApplicationJSON+"'")
		}
		inp := new(input)
		if err := c.Bind(inp); err != nil {
			goapp.Log.Error(err)
			return echo.NewHTTPError(http.StatusBadRequest, "Can get data")
		}

		res := &output{}
		res.Text = clean.Text(inp.Text)

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		c.Response().WriteHeader(http.StatusOK)
		enc := json.NewEncoder(c.Response())
		enc.SetEscapeHTML(false)
		return enc.Encode(res)
	}
}

func live(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		return c.JSONBlob(http.StatusOK, []byte(`{"service":"OK"}`))
	}
}
