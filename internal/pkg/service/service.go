package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

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
	//InfoGetter main sythesis processor
	InfoGetter interface {
		Provide(ID string) (*api.InfoResult, error)
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
		InfoGetterData InfoGetter
	}
)

//StartWebServer starts the HTTP service and listens for the admin requests
func StartWebServer(data *Data) error {
	goapp.Log.Infof("Starting HTTP TTS Line service at %d", data.Port)

	if err := validate(data); err != nil {
		return err
	}

	portStr := strconv.Itoa(data.Port)

	e := initRoutes(data)

	e.Server.Addr = ":" + portStr
	e.Server.IdleTimeout = 3 * time.Minute
	e.Server.ReadHeaderTimeout = 15 * time.Second
	e.Server.ReadTimeout = 60 * time.Second
	e.Server.WriteTimeout = 900 * time.Second

	w := goapp.Log.Writer()
	defer w.Close()
	gracehttp.SetLogger(log.New(w, "", 0))

	return gracehttp.Serve(e.Server)
}

func validate(data *Data) error {
	if data.InfoGetterData == nil {
		return errors.New("no infoGetter")
	}
	if data.SyntData.Processor == nil {
		return errors.New("no synt data")
	}
	if data.SyntCustomData.Processor == nil {
		return errors.New("no custom synt data")
	}
	return nil
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
	e.GET("/request/:requestID", synthesizeInfo(data.InfoGetterData))
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
			goapp.Log.Warn(err)
			return err
		}

		cfg, err := data.Configurator.Configure(c.Request(), inp)
		if err != nil {
			goapp.Log.Warn("Cannot prepare request config " + err.Error())
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		resp, err := data.Processor.Work(cfg)
		if err != nil {
			if d, msg := badReqError(err); d {
				goapp.Log.Warn("can't process: ", err)
				return echo.NewHTTPError(http.StatusBadRequest, msg)
			}
			goapp.Log.Error("can't process: ", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		return writeResponse(c, resp)
	}
}

func synthesizeCustom(data *PrData) func(echo.Context) error {
	return func(c echo.Context) error {
		defer goapp.Estimate("Service synthesize custom method")()

		rID := c.QueryParam("requestID")
		if rID == "" {
			goapp.Log.Warn("No requestID")
			return echo.NewHTTPError(http.StatusBadRequest, "No requestID")
		}

		inp, err := takeInput(c)
		if err != nil {
			goapp.Log.Warn(err)
			return err
		}

		if inp.AllowCollectData != nil && !*inp.AllowCollectData {
			goapp.Log.Warn("Can't call with inp.AllowCollectData=false")
			return echo.NewHTTPError(http.StatusBadRequest, "Method does not allow 'saveRequest=false'")
		}

		cfg, err := data.Configurator.Configure(c.Request(), inp)
		if err != nil {
			goapp.Log.Warn("Cannot prepare request config " + err.Error())
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		cfg.RequestID = rID
		cfg.AllowCollectData = true // turn collect data to true as it is mandatory for this request

		resp, err := data.Processor.Work(cfg)
		if err != nil {
			if d, msg := badReqError(err); d {
				goapp.Log.Warn("can't process: ", err)
				return echo.NewHTTPError(http.StatusBadRequest, msg)
			}
			goapp.Log.Error("can't process: ", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		resp.RequestID = ""

		return writeResponse(c, resp)
	}
}

func synthesizeInfo(data InfoGetter) func(echo.Context) error {
	return func(c echo.Context) error {
		defer goapp.Estimate("Service request info method")()

		rID := c.Param("requestID")
		if rID == "" {
			goapp.Log.Warn("No requestID")
			return echo.NewHTTPError(http.StatusBadRequest, "No requestID")
		}

		resp, err := data.Provide(rID)
		if err != nil {
			if d, msg := badReqError(err); d {
				goapp.Log.Warn("can't process. ", err)
				return echo.NewHTTPError(http.StatusBadRequest, msg)
			}
			goapp.Log.Error("can't process. ", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
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

func writeResponse(c echo.Context, resp interface{}) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
	c.Response().WriteHeader(http.StatusOK)
	enc := json.NewEncoder(c.Response())
	enc.SetEscapeHTML(false)
	return enc.Encode(resp)
}

func badReqError(err error) (bool, string) {
	if errors.Is(err, utils.ErrNoRecord) {
		return true, "RequestID not found"
	}
	if errors.Is(err, utils.ErrNoInput) {
		return true, "No text"
	}
	if errors.Is(err, utils.ErrTextDoesNotMatch) {
		return true, "Original text does not match the modified"
	}
	var errBA *utils.ErrBadAccent
	if errors.As(err, &errBA) {
		return true, fmt.Sprintf("Bad accents: %v", errBA.BadAccents)
	}
	var errWTL *utils.ErrWordTooLong
	if errors.As(err, &errWTL) {
		return true, fmt.Sprintf("Word too long: '%s'", errWTL.Word)
	}
	var errTTL *utils.ErrTextTooLong
	if errors.As(err, &errTTL) {
		return true, fmt.Sprintf("Text too long: passed %d chars, max allowed %d", errTTL.Len, errTTL.Max)
	}
	var errBadS *utils.ErrBadSymbols
	if errors.As(err, &errBadS) {
		return true, fmt.Sprintf("Wrong symbols: '%s'", errBadS.Orig)
	}
	return false, ""
}

func live(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		return c.JSONBlob(http.StatusOK, []byte(`{"service":"OK"}`))
	}
}
