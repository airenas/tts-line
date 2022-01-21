package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/exporter"
	"github.com/airenas/tts-line/internal/pkg/mongodb"
	"github.com/mattn/go-colorable"

	"github.com/labstack/gommon/color"
	"github.com/pkg/errors"
)

type params struct {
	delete bool
	to     time.Time
}

func main() {
	os.Setenv("LOGGER_OUT_NAME", "stderr")
	fs := flag.CommandLine
	ap := &params{}
	takeParams(fs, ap)
	goapp.StartWithFlags(fs, os.Args)

	goapp.Log.Info("Starting")

	printBanner()

	defer goapp.Estimate("Export")()

	sp, err := mongodb.NewSessionProvider(goapp.Config.GetString("mongo.url"))
	if err != nil {
		goapp.Log.Fatal(errors.Wrap(err, "can't init mongo session provider"))
	}
	defer sp.Close()
	ts, err := mongodb.NewTextSaver(sp)
	if err != nil {
		goapp.Log.Fatal(errors.Wrap(err, "can't init DB exporter"))
	}

	p := exporter.Params{To: ap.to, Delete: ap.delete, Out: os.Stdout, Exporter: ts}
	err = exporter.Export(p)
	if err != nil {
		goapp.Log.Fatal(errors.Wrap(err, "can't start the service"))
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

func takeParams(fs *flag.FlagSet, data *params) {
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage of %s: <params> > out.json\n", os.Args[0])
		fs.PrintDefaults()
	}
	fs.Var(timeValue{to: &data.to}, "to", "Filters records according to the time provided here. Takes older record than the time. Format 'YYYY-MM-DD'")
	fs.BoolVar(&data.delete, "delete", false, "Deletes filtered records from database")
}

type timeValue struct {
	to *time.Time
}

func (v timeValue) String() string {
	if v.to != nil && !v.to.IsZero() {
		return v.to.Format("2006-01-02")
	}
	return ""
}

func (v timeValue) Set(s string) error {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	*v.to = t
	return nil
}
