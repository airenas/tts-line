package service

import (
	"net/http"
	"strconv"
	"strings"

	"log"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/acronyms/service/api"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/pkg/errors"
)

type (
	// Worker returns acronyms as word to pronounce
	Worker interface {
		Process(word, mi string) ([]api.ResultWord, error)
	}

	//Data is service operation data
	Data struct {
		Worker Worker
		Port   int
	}
)

//StartWebServer starts the HTTP service and listens for the requests
func StartWebServer(data *Data) error {
	goapp.Log.Infof("Starting HTTP service at %d", data.Port)
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
	p := prometheus.NewPrometheus("acronyms", nil)
	p.Use(e)

	e.POST("/acronyms", handleList(data))
	e.GET("/acronym/:word", handleOne(data))
	e.GET("/live", live(data))

	goapp.Log.Info("Routes:")
	for _, r := range e.Routes() {
		goapp.Log.Infof("  %s %s", r.Method, r.Path)
	}
	return e
}

func handleOne(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		defer goapp.Estimate("Service method: abbreviation")()

		word := c.Param("word")
		if word == "" {
			goapp.Log.Error("No word")
			return echo.NewHTTPError(http.StatusBadRequest, "No word")
		}

		res, err := data.Worker.Process(word, "")
		if err != nil {
			goapp.Log.Error(errors.Wrap(err, "Cannot process "+word))
			return echo.NewHTTPError(http.StatusInternalServerError, "Cannot process "+word)
		}
		return c.JSON(http.StatusOK, res)
	}
}

func live(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		return c.JSONBlob(http.StatusOK, []byte(`{"service":"OK"}`))
	}
}

func handleList(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		defer goapp.Estimate("Service method: abbreviations")()

		ctype := c.Request().Header.Get(echo.HeaderContentType)
		if !strings.HasPrefix(ctype, echo.MIMEApplicationJSON) {
			goapp.Log.Error("Wrong content type")
			return echo.NewHTTPError(http.StatusBadRequest, "Wrong content type. Expected '"+echo.MIMEApplicationJSON+"'")
		}
		var input []api.WordInput
		if err := c.Bind(&input); err != nil {
			goapp.Log.Error(err)
			return echo.NewHTTPError(http.StatusBadRequest, "Can get data")
		}

		res := make([]*api.WordOutput, 0)
		for _, wi := range input {
			wl, err := data.Worker.Process(wi.Word, wi.MI)
			if err != nil {
				goapp.Log.Error("Cannot process "+wi.Word, err)
				return echo.NewHTTPError(http.StatusInternalServerError, "Cannot process "+wi.Word)
			}
			var wo api.WordOutput
			wo.ID = wi.ID
			wo.Words = wl
			res = append(res, &wo)
		}
		return c.JSON(http.StatusOK, res)
	}
}
