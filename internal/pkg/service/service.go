package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/utils"
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
	//PrData is method process data
	PrData struct {
		Processor    Synthesizer
		Configurator Configurator
	}

	//Data is service operation data
	Data struct {
		Port           int
		SyntData       PrData
		SyntCustomData PrData
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

var promMdlw *prometheus.Prometheus

func init() {
	promMdlw = prometheus.NewPrometheus("tts", nil)
}

func initRoutes(data *Data) *echo.Echo {
	e := echo.New()
	e.Use(middleware.Logger())
	promMdlw.Use(e)

	e.POST("/synthesize", synthesizeText(&data.SyntData))
	e.POST("/synthesizeCustom", synthesizeCustom(&data.SyntCustomData))
	e.GET("/live", live(data))

	goapp.Log.Info("Routes:")
	for _, r := range e.Routes() {
		goapp.Log.Infof("  %s %s", r.Method, r.Path)
	}
	return e
}

func synthesizeText(data *PrData) func(echo.Context) error {
	return func(c echo.Context) error {
		defer goapp.Estimate("Service synthesize method")()

		inp, err := takeInput(c)
		if err != nil {
			goapp.Log.Error(err)
			return err
		}

		cfg, err := data.Configurator.Configure(c.Request(), inp)
		if err != nil {
			goapp.Log.Error("Cannot prepare request config " + err.Error())
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		resp, err := data.Processor.Work(cfg)
		if err != nil {
			goapp.Log.Error("Can't process. ", err)
			if d, msg := badReqError(err); d {
				return echo.NewHTTPError(http.StatusBadRequest, msg)
			}
			return echo.NewHTTPError(http.StatusInternalServerError, "Service error")
		}

		return writeResponse(c, resp)
	}
}

func synthesizeCustom(data *PrData) func(echo.Context) error {
	return func(c echo.Context) error {
		defer goapp.Estimate("Service synthesize custom method")()

		rID := c.QueryParam("requestID")
		if rID == "" {
			goapp.Log.Error("No requestID")
			return echo.NewHTTPError(http.StatusBadRequest, "No requestID")
		}

		inp, err := takeInput(c)
		if err != nil {
			goapp.Log.Error(err)
			return err
		}

		if inp.AllowCollectData != nil && !*inp.AllowCollectData {
			goapp.Log.Error("Can call with inp.AllowCollectData=false")
			return echo.NewHTTPError(http.StatusBadRequest, "Method does not allow 'allowCollectData=false'")
		}

		cfg, err := data.Configurator.Configure(c.Request(), inp)
		if err != nil {
			goapp.Log.Error("Cannot prepare request config " + err.Error())
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		cfg.RequestID = rID
		cfg.AllowCollectData = true // turn collect data to true as it is mandatory for this request

		resp, err := data.Processor.Work(cfg)
		if err != nil {
			goapp.Log.Error("Can't process. ", err)
			if d, msg := badReqError(err); d {
				return echo.NewHTTPError(http.StatusBadRequest, msg)
			}
			return echo.NewHTTPError(http.StatusInternalServerError, "Service error")
		}
		resp.RequestID = ""

		return writeResponse(c, resp)
	}
}

func takeInput(c echo.Context) (*api.Input, error) {
	ctype := c.Request().Header.Get(echo.HeaderContentType)
	if !strings.HasPrefix(ctype, echo.MIMEApplicationJSON) {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Wrong content type. Expected '"+echo.MIMEApplicationJSON+"'")
	}
	inp := new(api.Input)
	if err := c.Bind(inp); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Cannot decode input")
	}
	return inp, nil
}

func writeResponse(c echo.Context, resp *api.Result) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
	c.Response().WriteHeader(getCode(resp))
	enc := json.NewEncoder(c.Response())
	enc.SetEscapeHTML(false)
	return enc.Encode(resp)
}

func getCode(resp *api.Result) int {
	if len(resp.ValidationFailures) > 0 {
		return http.StatusBadRequest
	}
	return 200
}

func badReqError(err error) (bool, string) {
	if errors.Is(err, utils.ErrNoRecord) {
		return true, "RequestID not found"
	}
	if errors.Is(err, utils.ErrTextDoesNotMatch) {
		return true, "Original text does not match the modified"
	}
	var errBA *utils.ErrBadAccent
	if errors.As(err, &errBA) {
		return true, fmt.Sprintf("Bad accents: %v", errBA.BadAccents)
	}
	var errWTA *utils.ErrWordTooLong
	if errors.As(err, &errWTA) {
		return true, fmt.Sprintf("Word too long: '%s'", errWTA.Word)
	}
	return false, ""
}

func live(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		return c.JSONBlob(http.StatusOK, []byte(`{"service":"OK"}`))
	}
}
