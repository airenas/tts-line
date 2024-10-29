package main

import (
	"os"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/acronyms"
	"github.com/airenas/tts-line/internal/pkg/acronyms/service"
	"github.com/labstack/gommon/color"

	"github.com/pkg/errors"
)

func main() {
	goapp.StartWithDefault()

	data := service.Data{}
	data.Port = goapp.Config.GetInt("port")

	var err error
	data.Worker, err = provider()
	if err != nil {
		goapp.Log.Fatal().Err(errors.Wrap(err, "Can't init provider")).Send()
	}

	printBanner()

	err = service.StartWebServer(&data)
	if err != nil {
		goapp.Log.Fatal().Err(errors.Wrap(err, "Can't start the service")).Send()
	}
}

func provider() (service.Worker, error) {
	ls, err := acronyms.NewLetters()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to init letters")
	}

	fStr := goapp.Config.GetString("acronyms.file")
	fileA, err := os.Open(fStr)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to open file: "+fStr)
	}
	defer fileA.Close()
	as, err := acronyms.New(fileA)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to open file: "+fStr)
	}

	return acronyms.NewProcessor(as, ls)
}

var (
	version string
)

func printBanner() {
	banner := `
    ___                                             
   /   | ______________  ____  __  ______ ___  _____
  / /| |/ ___/ ___/ __ \/ __ \/ / / / __ ` + "`" + `__ \/ ___/
 / ___ / /__/ /  / /_/ / / / / /_/ / / / / / (__  ) 
/_/  |_\___/_/   \____/_/ /_/\__, /_/ /_/ /_/____/  
              ________  ____//___/_(_)_______       
             / ___/ _ \/ ___/ | / / / ___/ _ \      
            (__  )  __/ /   | |/ / / /__/  __/      
           /____/\___/_/    |___/_/\___/\___/  v: %s 

%s
________________________________________________________                                                 

`
	cl := color.New()
	cl.Printf(banner, cl.Red(version), cl.Green("https://github.com/airenas/tts-line"))
}
