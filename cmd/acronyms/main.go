package main

import (
	"os"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/acronyms"
	"github.com/airenas/tts-line/internal/pkg/acronyms/service"

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

	err = service.StartWebServer(&data)
	if err != nil {
		goapp.Log.Fatal(errors.Wrap(err, "Can't start the service"))
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
