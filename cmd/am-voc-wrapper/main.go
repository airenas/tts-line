package main

import (
	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/wrapservice"

	"github.com/pkg/errors"
)

func main() {
	goapp.StartWithDefault()

	data := wrapservice.Data{}
	data.Port = goapp.Config.GetInt("port")
	var err error
	data.Processor, err = wrapservice.NewProcessor(goapp.Config.GetString("acousticModel.url"), goapp.Config.GetString("vocoder.url"))
	if err != nil {
		goapp.Log.Fatal(errors.Wrap(err, "Can't init processor"))
	}

	err = wrapservice.StartWebServer(&data)
	if err != nil {
		goapp.Log.Fatal(errors.Wrap(err, "Can't start the service"))
	}
}
