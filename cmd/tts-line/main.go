package main

import (
	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/processor"
	"github.com/airenas/tts-line/internal/pkg/service"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"

	"github.com/pkg/errors"
)

func main() {
	goapp.StartWithDefault()

	data := service.Data{}
	data.Port = goapp.Config.GetInt("port")
	synt := &synthesizer.MainWorker{}
	err := addProcessors(synt)
	if err != nil {
		goapp.Log.Fatal(errors.Wrap(err, "Can't init processors"))
	}
	data.Processor = synt
	err = service.StartWebServer(&data)
	if err != nil {
		goapp.Log.Fatal(errors.Wrap(err, "Can't start the service"))
	}
}

func addProcessors(synt *synthesizer.MainWorker) error {
	synt.Processors = append(synt.Processors, processor.NewNormalizer())
	return nil
}
