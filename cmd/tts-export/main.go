package main

import (
	"os"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/exporter"
	"github.com/airenas/tts-line/internal/pkg/mongodb"
	"github.com/mattn/go-colorable"

	"github.com/labstack/gommon/color"
	"github.com/pkg/errors"
)

func main() {
	os.Setenv("LOGGER_OUT_NAME", "stderr")
	goapp.StartWithDefault()
	goapp.Log.Info("Starting")

	printBanner()

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

var (
	version string
)

func printBanner() {
	banner := `
  _________________                               __ 
 /_  __/_  __/ ___/   ___  _  ______  ____  _____/ /_
  / /   / /  \__ \   / _ \| |/_/ __ \/ __ \/ ___/ __/
 / /   / /  ___/ /  /  __/>  </ /_/ / /_/ / /  / /_  
/_/   /_/  /____/   \___/_/|_/ .___/\____/_/   \__/  v: %s
							   /_/                      
   
%s
________________________________________________________                                                 

`
	cl := color.New()
	cl.SetOutput(colorable.NewColorableStderr())
	cl.Printf(banner, cl.Red(version), cl.Green("https://github.com/airenas/tts-line"))
}
