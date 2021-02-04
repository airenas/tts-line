package main

import (
	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/clean/service"

	"github.com/pkg/errors"
)

func main() {
	goapp.StartWithDefault()

	data := service.Data{}
	data.Port = goapp.Config.GetInt("port")
	err := service.StartWebServer(&data)
	if err != nil {
		goapp.Log.Fatal(errors.Wrap(err, "Can't start the service"))
	}
}
