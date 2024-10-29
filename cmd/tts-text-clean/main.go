package main

import (
	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/clean/service"
	"github.com/labstack/gommon/color"

	"github.com/pkg/errors"
)

func main() {
	goapp.StartWithDefault()

	data := service.Data{}
	data.Port = goapp.Config.GetInt("port")
	printBanner()
	err := service.StartWebServer(&data)
	if err != nil {
		goapp.Log.Fatal().Err(errors.Wrap(err, "Can't start the service")).Send()
	}
}

var (
	version string
)

func printBanner() {
	banner := `
  _________________    __            __ 
 /_  __/_  __/ ___/   / /____  _  __/ /_
  / /   / /  \__ \   / __/ _ \| |/_/ __/
 / /   / /  ___/ /  / /_/  __/>  </ /_  
/_/   /_/  /____/   \__/\___/_/|_|\__/  
          __                          
    _____/ /__  ____ _____  ___  _____
   / ___/ / _ \/ __ ` + "`" + `/ __ \/ _ \/ ___/
  / /__/ /  __/ /_/ / / / /  __/ /    
  \___/_/\___/\__,_/_/ /_/\___/_/  v: %s 

%s
________________________________________________________                                                 

`
	cl := color.New()
	cl.Printf(banner, cl.Red(version), cl.Green("https://github.com/airenas/tts-line"))
}
