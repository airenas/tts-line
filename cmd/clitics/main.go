package main

import (
	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/clitics"
	"github.com/airenas/tts-line/internal/pkg/clitics/service"
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
		goapp.Log.Fatal(errors.Wrap(err, "Can't init provider"))
	}

	printBanner()

	err = service.StartWebServer(&data)
	if err != nil {
		goapp.Log.Fatal(errors.Wrap(err, "Can't start the service"))
	}
}

func provider() (service.ClitWorker, error) {
	cl, err := clitics.ReadClitics(goapp.Config.GetString("clitics.file"))
	if err != nil {
		return nil, errors.Wrap(err, "unable to read clitics")
	}
	return clitics.NewProcessor(cl)
}

var (
	version string
)

func printBanner() {
	banner := `
             _________ __  _          
            / ____/ (_) /_(_)_________
           / /   / / / __/ / ___/ ___/
          / /___/ / / /_/ / /__(__  ) 
          \____/_/_/\__/_/\___/____/  
                            _         
      ________  ______   __(_)_______ 
     / ___/ _ \/ ___/ | / / / ___/ _ \
    (__  )  __/ /   | |/ / / /__/  __/
   /____/\___/_/    |___/_/\___/\___/  v: %s 

%s
________________________________________________________                                                 

`
	cl := color.New()
	cl.Printf(banner, cl.Red(version), cl.Green("https://github.com/airenas/tts-line"))
}
