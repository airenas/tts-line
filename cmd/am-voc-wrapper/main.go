package main

import (
	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/wrapservice"
	"github.com/labstack/gommon/color"

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

	printBanner()

	err = wrapservice.StartWebServer(&data)
	if err != nil {
		goapp.Log.Fatal(errors.Wrap(err, "Can't start the service"))
	}
}

var (
	version string
)

func printBanner() {
	banner := `
    ___    __  ___            _    ______  ______
   /   |  /  |/  /           | |  / / __ \/ ____/
  / /| | / /|_/ /  ______    | | / / / / / /     
 / ___ |/ /  / /  /_____/    | |/ / /_/ / /___   
/_/  |_/_/  /_/              |___/\____/\____/   
                                                 
 _       ______  ___    ____  ____  __________ 
| |     / / __ \/   |  / __ \/ __ \/ ____/ __ \
| | /| / / /_/ / /| | / /_/ / /_/ / __/ / /_/ /
| |/ |/ / _, _/ ___ |/ ____/ ____/ /___/ _, _/ 
|__/|__/_/ |_/_/  |_/_/   /_/   /_____/_/ |_|   v: %s 

%s
________________________________________________________                                                 

`
	cl := color.New()
	cl.Printf(banner, cl.Red(version), cl.Green("https://github.com/airenas/tts-line"))
}
