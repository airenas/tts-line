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
	pr, err := processor.NewNumberReplace(goapp.Config.GetString("numberReplace.url"))
	if err != nil {
		return errors.Wrap(err, "Can't init number replace")
	}
	synt.Processors = append(synt.Processors, pr)

	pr, err = processor.NewTagger(goapp.Config.GetString("tagger.url"))
	if err != nil {
		return errors.Wrap(err, "Can't init tagger")
	}
	synt.Processors = append(synt.Processors, pr)

	pr, err = processor.NewValidator(goapp.Sub(goapp.Config, "validator"))
	if err != nil {
		return errors.Wrap(err, "Can't init validator")
	}
	synt.Processors = append(synt.Processors, pr)

	pr, err = processor.NewAbbreviator(goapp.Config.GetString("abbreviator.url"))
	if err != nil {
		return errors.Wrap(err, "Can't init abbreviator")
	}
	synt.Processors = append(synt.Processors, pr)
	return nil
}
