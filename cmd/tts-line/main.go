package main

import (
	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/cache"
	"github.com/airenas/tts-line/internal/pkg/processor"
	"github.com/airenas/tts-line/internal/pkg/service"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"

	"github.com/pkg/errors"
)

func main() {
	goapp.StartWithDefault()

	data := service.Data{}
	data.Port = goapp.Config.GetInt("port")
	utils.MaxLogDataSize = goapp.Config.GetInt("maxLogDataSize")
	synt := &synthesizer.MainWorker{}
	err := addProcessors(synt)
	if err != nil {
		goapp.Log.Fatal(errors.Wrap(err, "Can't init processors"))
	}
	cc := goapp.Sub(goapp.Config, "cache")
	if cc != nil {
		data.Processor, err = cache.NewCacher(synt, cc)
		if err != nil {
			goapp.Log.Fatal(errors.Wrap(err, "Can't init cache"))
		}
	} else {
		goapp.Log.Info("No cache will be used")
		data.Processor = synt
	}
	err = service.StartWebServer(&data)
	if err != nil {
		goapp.Log.Fatal(errors.Wrap(err, "Can't start the service"))
	}
}

func addProcessors(synt *synthesizer.MainWorker) error {
	synt.Add(processor.NewNormalizer())
	pr, err := processor.NewNumberReplace(goapp.Config.GetString("numberReplace.url"))
	if err != nil {
		return errors.Wrap(err, "Can't init number replace")
	}
	synt.Add(pr)

	pr, err = processor.NewTagger(goapp.Config.GetString("tagger.url"))
	if err != nil {
		return errors.Wrap(err, "Can't init tagger")
	}
	synt.Add(pr)

	pr, err = processor.NewValidator(goapp.Sub(goapp.Config, "validator"))
	if err != nil {
		return errors.Wrap(err, "Can't init validator")
	}
	synt.Add(pr)

	synt.Add(processor.NewSplitter())

	partRunner := synthesizer.NewPartRunner()
	synt.Add(partRunner)

	synt.Add(processor.NewJoinAudio())

	pr, err = processor.NewMP3(goapp.Config.GetString("mp3.url"))
	if err != nil {
		return errors.Wrap(err, "Can't init mp3 converter")
	}
	synt.Add(pr)

	if goapp.Config.GetString("filer.dir") != "" {
		pr, err = processor.NewFiler(goapp.Config.GetString("filer.dir"))
		if err != nil {
			return errors.Wrap(err, "Can't init filer")
		}
		synt.Add(pr)
	}

	ppr, err := processor.NewAbbreviator(goapp.Config.GetString("abbreviator.url"))
	if err != nil {
		return errors.Wrap(err, "Can't init abbreviator")
	}
	partRunner.Add(ppr)

	ppr, err = processor.NewAccentuator(goapp.Config.GetString("accenter.url"))
	if err != nil {
		return errors.Wrap(err, "Can't init accenter")
	}
	partRunner.Add(ppr)

	ppr, err = processor.NewTranscriber(goapp.Config.GetString("transcriber.url"))
	if err != nil {
		return errors.Wrap(err, "Can't init transcriber")
	}
	partRunner.Add(ppr)

	ppr, err = processor.NewAcousticModel(goapp.Config.GetString("acousticModel.url"),
		goapp.Config.GetString("acousticModel.spaceSymbol"))
	if err != nil {
		return errors.Wrap(err, "Can't init acousticModel")
	}
	partRunner.Add(ppr)

	ppr, err = processor.NewVocoder(goapp.Config.GetString("vocoder.url"))
	if err != nil {
		return errors.Wrap(err, "Can't init vocoder")
	}
	partRunner.Add(ppr)

	return nil
}
