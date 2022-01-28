package main

import (
	"strconv"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/cache"
	"github.com/airenas/tts-line/internal/pkg/mongodb"
	"github.com/airenas/tts-line/internal/pkg/processor"
	"github.com/airenas/tts-line/internal/pkg/service"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/labstack/gommon/color"
	"github.com/spf13/viper"

	"net/http"
	_ "net/http/pprof"

	"github.com/pkg/errors"
)

func main() {
	goapp.StartWithDefault()

	data := service.Data{}
	data.Port = goapp.Config.GetInt("port")
	utils.MaxLogDataSize = goapp.Config.GetInt("maxLogDataSize")
	synt := &synthesizer.MainWorker{}
	synt.AllowCustomCode = goapp.Config.GetBool("allowCustom")
	sp, err := mongodb.NewSessionProvider(goapp.Config.GetString("mongo.url"))
	if err != nil {
		goapp.Log.Fatal(errors.Wrap(err, "can't init mongo session provider"))
	}
	defer sp.Close()

	err = addProcessors(synt, sp)
	if err != nil {
		goapp.Log.Fatal(errors.Wrap(err, "can't init processors"))
	}

	//cache
	cc := goapp.Sub(goapp.Config, "cache")
	if cc != nil {
		data.SyntData.Processor, err = cache.NewCacher(synt, cc)
		if err != nil {
			goapp.Log.Fatal(errors.Wrap(err, "can't init cache"))
		}
	} else {
		goapp.Log.Info("No cache will be used")
		data.SyntData.Processor = synt
	}

	// input configuration
	data.SyntData.Configurator, err = service.NewTTSConfigurator(goapp.Sub(goapp.Config, "options"))
	if err != nil {
		goapp.Log.Fatal(errors.Wrap(err, "can't init configurator"))
	}

	// init custom synthesize method
	data.SyntCustomData.Configurator, err = service.NewTTSConfigurator(goapp.Sub(goapp.Config, "options"))
	if err != nil {
		goapp.Log.Fatal(errors.Wrap(err, "can't init custom configurator"))
	}
	syntC := &synthesizer.MainWorker{}
	err = addCustomProcessors(syntC, sp, goapp.Config)
	if err != nil {
		goapp.Log.Fatal(errors.Wrap(err, "can't init custom processors"))
	}
	data.SyntCustomData.Processor = syntC

	printBanner()

	go startPerfEndpoint()

	err = service.StartWebServer(&data)
	if err != nil {
		goapp.Log.Fatal(errors.Wrap(err, "can't start the service"))
	}
}

func addProcessors(synt *synthesizer.MainWorker, sp *mongodb.SessionProvider) error {
	pr, err := processor.NewAddMetrics(processor.NewMetricsCharsFunc("/synthesize"))
	if err != nil {
		return errors.Wrap(err, "can't init metrics processor")
	}
	synt.Add(pr)
	//validator
	pr, err = processor.NewValidator(goapp.Config.GetInt("validator.maxChars"))
	if err != nil {
		return errors.Wrap(err, "can't init validator")
	}
	synt.Add(pr)
	
	ts, err := mongodb.NewTextSaver(sp)
	if err != nil {
		return errors.Wrap(err, "can't init text to DB saver")
	}
	sv, err := processor.NewSaver(ts, utils.RequestOriginal)
	if err != nil {
		return errors.Wrap(err, "can't init text to DB saver")
	}
	synt.Add(sv)
	// cleaner
	pr, err = processor.NewCleaner(goapp.Config.GetString("clean.url"))
	if err != nil {
		return errors.Wrap(err, "can't init normalize/clean processor")
	}
	synt.Add(pr)
	
	synt.Add(processor.NewURLReplacer())
	//db saver
	sv, err = processor.NewSaver(ts, utils.RequestCleaned)
	if err != nil {
		return errors.Wrap(err, "can't init text to DB saver")
	}
	synt.Add(sv)
	//number replacer
	pr, err = processor.NewNumberReplace(goapp.Config.GetString("numberReplace.url"))
	if err != nil {
		return errors.Wrap(err, "can't init number replace")
	}
	synt.Add(pr)
	//db saver
	sv, err = processor.NewSaver(ts, utils.RequestNormalized)
	if err != nil {
		return errors.Wrap(err, "can't init text to DB saver")
	}
	synt.Add(sv)

	pr, err = processor.NewTagger(goapp.Config.GetString("tagger.url"))
	if err != nil {
		return errors.Wrap(err, "can't init tagger")
	}
	synt.Add(pr)

	synt.Add(processor.NewSplitter(goapp.Config.GetInt("splitter.maxChars")))

	partRunner := synthesizer.NewPartRunner(goapp.Config.GetInt("partRunner.workers"))
	synt.Add(partRunner)

	synt.Add(processor.NewJoinAudio())

	pr, err = processor.NewConverter(goapp.Config.GetString("audioConvert.url"))
	if err != nil {
		return errors.Wrap(err, "can't init mp3 converter")
	}
	synt.Add(pr)

	pr, err = processor.NewAddMetrics(processor.NewMetricsWaveLenFunc("/synthesize"))
	if err != nil {
		return errors.Wrap(err, "can't init metrics processor")
	}
	synt.Add(pr)

	if goapp.Config.GetString("filer.dir") != "" {
		pr, err = processor.NewFiler(goapp.Config.GetString("filer.dir"))
		if err != nil {
			return errors.Wrap(err, "can't init filer")
		}
		synt.Add(pr)
	}
	return addPartProcessors(partRunner, goapp.Config)
}

func addCustomProcessors(synt *synthesizer.MainWorker, sp *mongodb.SessionProvider, cfg *viper.Viper) error {
	pr, err := processor.NewAddMetrics(processor.NewMetricsCharsFunc("/synthesizeCustom"))
	if err != nil {
		return errors.Wrap(err, "can't init metrics processor")
	}
	synt.Add(pr)

	pr, err = processor.NewValidator(cfg.GetInt("validator.maxChars"))
	if err != nil {
		return errors.Wrap(err, "can't init validator")
	}
	synt.Add(pr)

	ts, err := mongodb.NewTextSaver(sp)
	if err != nil {
		return errors.Wrap(err, "can't init text to DB saver")
	}

	pr, err = processor.NewLoader(ts)
	if err != nil {
		return errors.Wrap(err, "can't init text from DB loader")
	}
	synt.Add(pr)


	pr, err = processor.NewComparator(cfg.GetString("comparator.url"))
	if err != nil {
		return errors.Wrap(err, "can't init text comparator")
	}
	synt.Add(pr)

	sv, err := processor.NewSaver(ts, utils.RequestUser)
	if err != nil {
		return errors.Wrap(err, "can't init text to DB saver")
	}
	synt.Add(sv)

	pr, err = processor.NewTaggerAccents(cfg.GetString("tagger.url"))
	if err != nil {
		return errors.Wrap(err, "can't init tagger")
	}
	synt.Add(pr)

	synt.Add(processor.NewSplitter(cfg.GetInt("splitter.maxChars")))

	partRunner := synthesizer.NewPartRunner(cfg.GetInt("partRunner.workers"))
	synt.Add(partRunner)

	synt.Add(processor.NewJoinAudio())

	pr, err = processor.NewConverter(cfg.GetString("audioConvert.url"))
	if err != nil {
		return errors.Wrap(err, "can't init audioConvert converter")
	}
	synt.Add(pr)

	pr, err = processor.NewAddMetrics(processor.NewMetricsWaveLenFunc("/synthesizeCustom"))
	if err != nil {
		return errors.Wrap(err, "can't init metrics processor")
	}
	synt.Add(pr)

	if goapp.Config.GetString("filer.dir") != "" {
		pr, err = processor.NewFiler(cfg.GetString("filer.dir"))
		if err != nil {
			return errors.Wrap(err, "can't init filer")
		}
		synt.Add(pr)
	}
	return addPartProcessors(partRunner, cfg)
}

func addPartProcessors(partRunner *synthesizer.PartRunner, cfg *viper.Viper) error {
	ppr, err := processor.NewObsceneFilter(cfg.GetString("obscene.url"))
	if err != nil {
		return errors.Wrap(err, "can't init obscene filter service")
	}
	partRunner.Add(ppr)

	ppr, err = processor.NewAcronyms(cfg.GetString("acronyms.url"))
	if err != nil {
		return errors.Wrap(err, "can't init acronyms service")
	}
	partRunner.Add(ppr)

	ppr, err = processor.NewAccentuator(cfg.GetString("accenter.url"))
	if err != nil {
		return errors.Wrap(err, "can't init accenter")
	}
	partRunner.Add(ppr)

	ppr, err = processor.NewClitics(cfg.GetString("clitics.url"))
	if err != nil {
		return errors.Wrap(err, "can't init clitics")
	}
	partRunner.Add(ppr)

	ppr, err = processor.NewTranscriber(cfg.GetString("transcriber.url"))
	if err != nil {
		return errors.Wrap(err, "can't init transcriber")
	}
	partRunner.Add(ppr)

	ppr, err = processor.NewAcousticModel(goapp.Sub(cfg, "acousticModel"))
	if err != nil {
		return errors.Wrap(err, "can't init acousticModel")
	}
	partRunner.Add(ppr)

	if !cfg.GetBool("acousticModel.hasVocoder") {
		ppr, err = processor.NewVocoder(cfg.GetString("vocoder.url"))
		if err != nil {
			return errors.Wrap(err, "can't init vocoder")
		}
		partRunner.Add(ppr)
	}

	return nil
}

func startPerfEndpoint() {
	port := goapp.Config.GetInt("debug.port")
	if port > 0 {
		goapp.Log.Infof("Starting Debug http endpoit at [::]:%d", port)
		portStr := strconv.Itoa(port)
		err := http.ListenAndServe(":"+portStr, nil)
		if err != nil {
			goapp.Log.Error(errors.Wrap(err, "can't start Debug endpoint at "+portStr))
		}
	}
}

var (
	version string
)

func printBanner() {
	banner := `
  _________________    ___          
 /_  __/_  __/ ___/   / (_)___  ___ 
  / /   / /  \__ \   / / / __ \/ _ \
 / /   / /  ___/ /  / / / / / /  __/
/_/   /_/  /____/  /_/_/_/ /_/\___/  v: %s 

%s
________________________________________________________                                                 

`
	cl := color.New()
	cl.Printf(banner, cl.Red(version), cl.Green("https://github.com/airenas/tts-line"))
}
