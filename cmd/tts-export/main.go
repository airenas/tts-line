package main

import (
	"os"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/exporter"
	"github.com/airenas/tts-line/internal/pkg/mongodb"

	"github.com/pkg/errors"
)

func main() {
	os.Setenv("LOGGER_OUT_NAME", "stderr")
	goapp.StartWithDefault()
	goapp.Log.Info("Starting")

	defer goapp.Estimate("Export")()

	sp, err := mongodb.NewSessionProvider(goapp.Config.GetString("mongo.url"))
	if err != nil {
		goapp.Log.Fatal(errors.Wrap(err, "Can't init mongo session provider"))
	}
	defer sp.Close()
	ts, err := mongodb.NewTextSaver(sp)
	if err != nil {
		goapp.Log.Fatal(errors.Wrap(err, "Can't init DB exporter"))
	}

	err = exporter.Export(ts, os.Stdout)
	if err != nil {
		goapp.Log.Fatal(errors.Wrap(err, "Can't start the service"))
	}
	goapp.Log.Info("Finished")
}
