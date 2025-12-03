package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/cache"
	"github.com/airenas/tts-line/internal/pkg/file"
	"github.com/airenas/tts-line/internal/pkg/mongodb"
	"github.com/airenas/tts-line/internal/pkg/processor"
	"github.com/airenas/tts-line/internal/pkg/service"
	sapi "github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/labstack/gommon/color"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"

	"net/http"
	_ "net/http/pprof"

	"github.com/pkg/errors"
)

func main() {
	goapp.StartWithDefault()
	log.Logger = goapp.Log
	zerolog.DefaultContextLogger = &goapp.Log

	if err := mainInt(context.Background()); err != nil {
		log.Fatal().Err(err).Send()
	}
}

func mainInt(ctx context.Context) error {
	tp, err := initTracer(ctx, goapp.Config.GetString("otel.exporter.otlp.endpoint"))
	if err != nil {
		return fmt.Errorf("init tracer: %w", err)
	}
	if s, ok := tp.(shutdowner); ok {
		defer func() {
			ctx, cf := context.WithTimeout(context.Background(), time.Second*5)
			defer cf()
			err := s.Shutdown(ctx)
			if err != nil {
				log.Warn().Err(err).Msg("Failed to shutdown OpenTelemetry")
			}
		}()
	}

	data := service.Data{}
	data.Port = goapp.Config.GetInt("port")
	utils.MaxLogDataSize = goapp.Config.GetInt("maxLogDataSize")
	synt := &synthesizer.MainWorker{}
	synt.AllowCustomCode = goapp.Config.GetBool("allowCustom")
	sp, err := mongodb.NewSessionProvider(goapp.Config.GetString("mongo.url"))
	if err != nil {
		return fmt.Errorf("init mongo session provider: %w", err)
	}
	defer sp.Close()

	if err = addProcessors(synt, sp, goapp.Config); err != nil {
		return fmt.Errorf("init processors: %w", err)
	}

	if err = addSSMLProcessors(synt, sp, goapp.Config); err != nil {
		return fmt.Errorf("init SSML processors: %w", err)
	}

	//cache
	cc := goapp.Sub(goapp.Config, "cache")
	if cc != nil {
		data.SyntData.Processor, err = cache.NewCacher(synt, cc)
		if err != nil {
			return fmt.Errorf("init cache: %w", err)
		}
	} else {
		goapp.Log.Info().Msg("No cache will be used")
		data.SyntData.Processor = synt
	}

	// input configuration
	data.SyntData.Configurator, err = service.NewTTSConfigurator(goapp.Sub(goapp.Config, "options"))
	if err != nil {
		return fmt.Errorf("init configurator: %w", err)
	}

	// init custom synthesize method
	data.SyntCustomData.Configurator, err = service.NewTTSConfiguratorNoSSML(goapp.Sub(goapp.Config, "options"))
	if err != nil {
		return fmt.Errorf("init custom configurator: %w", err)
	}
	syntC := &synthesizer.MainWorker{}
	err = addCustomProcessors(syntC, sp, goapp.Config)
	if err != nil {
		return fmt.Errorf("init custom processors: %w", err)
	}
	data.SyntCustomData.Processor = syntC
	data.InfoGetterData, err = prepareInfoGetter(sp)
	if err != nil {
		return fmt.Errorf("init info getter: %w", err)
	}
	printBanner()

	go startPerfEndpoint()

	err = service.StartWebServer(&data)
	if err != nil {
		return fmt.Errorf("start the service: %w", err)
	}
	return nil
}

func addProcessors(synt *synthesizer.MainWorker, sp *mongodb.SessionProvider, cfg *viper.Viper) error {
	pr, err := processor.NewAddMetrics(processor.NewMetricsCharsFunc("/synthesize"))
	if err != nil {
		return errors.Wrap(err, "can't init metrics processor")
	}
	synt.Add(pr)
	//validator
	pr, err = processor.NewValidator(cfg.GetInt("validator.maxChars"))
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
	pr, err = processor.NewCleaner(cfg.GetString("clean.url"))
	if err != nil {
		return errors.Wrap(err, "can't init normalize/clean processor")
	}
	synt.Add(pr)

	// normalizer
	pr, err = processor.NewNormalizer(cfg.GetString("normalize.url"))
	if err != nil {
		return errors.Wrap(err, "can't init text normalizer processor")
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
	pr, err = processor.NewNumberReplace(cfg.GetString("numberReplace.url"))
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

	pr, err = processor.NewTagger(cfg.GetString("tagger.url"))
	if err != nil {
		return errors.Wrap(err, "can't init tagger")
	}
	synt.Add(pr)

	pr, err = processor.NewTransliterator(cfg.GetString("transliterator.url"))
	if err != nil {
		return fmt.Errorf("can't init transliterator: %w", err)
	}
	synt.Add(pr)

	synt.Add(processor.NewSplitter(cfg.GetInt("splitter.maxChars")))

	partRunner := synthesizer.NewPartRunner(cfg.GetInt("partRunner.workers"))
	synt.Add(partRunner)

	suffixLoader, err := file.NewLoader(cfg.GetString("suffixLoader.path"))
	if err != nil {
		return errors.Wrap(err, "can't init suffix Loader")
	}

	synt.Add(processor.NewCalcLoudness())
	synt.Add(processor.NewJoinAudio(suffixLoader))

	pr, err = processor.NewConverter(cfg.GetString("audioConvert.url"))
	if err != nil {
		return errors.Wrap(err, "can't init mp3 converter")
	}
	synt.Add(pr)

	pr, err = processor.NewAddMetrics(processor.NewMetricsWaveLenFunc("/synthesize"))
	if err != nil {
		return errors.Wrap(err, "can't init metrics processor")
	}
	synt.Add(pr)

	if cfg.GetString("filer.dir") != "" {
		pr, err = processor.NewFiler(cfg.GetString("filer.dir"))
		if err != nil {
			return errors.Wrap(err, "can't init filer")
		}
		synt.Add(pr)
	}
	return addPartProcessors(partRunner, cfg)
}

func addSSMLProcessors(synt *synthesizer.MainWorker, sp *mongodb.SessionProvider, cfg *viper.Viper) error {
	pr, err := processor.NewAddMetrics(processor.NewMetricsCharsFunc("/synthesize"))
	if err != nil {
		return errors.Wrap(err, "can't init metrics processor")
	}
	synt.AddSSML(pr)
	//validator
	pr, err = processor.NewSSMLValidator(cfg.GetInt("validator.maxChars"))
	if err != nil {
		return errors.Wrap(err, "can't init validator")
	}
	synt.AddSSML(pr)

	ts, err := mongodb.NewTextSaver(sp)
	if err != nil {
		return errors.Wrap(err, "can't init text to DB saver")
	}
	sv, err := processor.NewSaver(ts, utils.RequestOriginalSSML)
	if err != nil {
		return errors.Wrap(err, "can't init text to DB saver")
	}
	synt.AddSSML(sv)

	var processors []synthesizer.Processor

	// cleaner
	pr, err = processor.NewCleaner(cfg.GetString("clean.url"))
	if err != nil {
		return errors.Wrap(err, "can't init normalize/clean processor")
	}
	processors = append(processors, pr)
	// normalizer
	pr, err = processor.NewNormalizer(cfg.GetString("normalize.url"))
	if err != nil {
		return errors.Wrap(err, "can't init text normalizer processor")
	}
	processors = append(processors, pr)
	processors = append(processors, processor.NewURLReplacer())
	//number replacer
	pr, err = processor.NewSSMLNumberReplace(cfg.GetString("numberReplace.url"))
	if err != nil {
		return errors.Wrap(err, "can't init number replace")
	}
	processors = append(processors, pr)
	pr, err = processor.NewSSMLTagger(cfg.GetString("tagger.url"))
	if err != nil {
		return errors.Wrap(err, "can't init tagger")
	}
	processors = append(processors, pr)

	pr, err = processor.NewTransliterator(cfg.GetString("transliterator.url"))
	if err != nil {
		return fmt.Errorf("can't init transliterator: %w", err)
	}
	processors = append(processors, pr)

	processors = append(processors, processor.NewSplitter(cfg.GetInt("splitter.maxChars")))

	partRunner := synthesizer.NewPartRunner(cfg.GetInt("partRunner.workers"))
	processors = append(processors, partRunner)

	synt.AddSSML(processor.NewSSMLPartRunner(processors))

	suffixLoader, err := file.NewLoader(cfg.GetString("suffixLoader.path"))
	if err != nil {
		return errors.Wrap(err, "can't init suffix Loader")
	}
	synt.AddSSML(processor.NewCalcLoudnessSSML())
	synt.AddSSML(processor.NewJoinSSMLAudio(suffixLoader))

	pr, err = processor.NewConverter(cfg.GetString("audioConvert.url"))
	if err != nil {
		return errors.Wrap(err, "can't init mp3 converter")
	}
	synt.AddSSML(pr)

	pr, err = processor.NewAddMetrics(processor.NewMetricsWaveLenFunc("/synthesize"))
	if err != nil {
		return errors.Wrap(err, "can't init metrics processor")
	}
	synt.AddSSML(pr)

	if cfg.GetString("filer.dir") != "" {
		pr, err = processor.NewFiler(cfg.GetString("filer.dir"))
		if err != nil {
			return errors.Wrap(err, "can't init filer")
		}
		synt.AddSSML(pr)
	}
	return addPartProcessors(partRunner, cfg)
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

	pr, err = processor.NewTransliterator(cfg.GetString("transliterator.url"))
	if err != nil {
		return fmt.Errorf("can't init transliterator: %w", err)
	}
	synt.Add(pr)

	synt.Add(processor.NewSplitter(cfg.GetInt("splitter.maxChars")))

	partRunner := synthesizer.NewPartRunner(cfg.GetInt("partRunner.workers"))
	synt.Add(partRunner)

	suffixLoader, err := file.NewLoader(cfg.GetString("suffixLoader.path"))
	if err != nil {
		return errors.Wrap(err, "can't init suffix Loader")
	}
	synt.Add(processor.NewCalcLoudness())
	synt.Add(processor.NewJoinAudio(suffixLoader))

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

	if cfg.GetString("filer.dir") != "" {
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

type infoGetter struct {
	ts *mongodb.TextSaver
}

func (ig *infoGetter) Provide(rID string) (*sapi.InfoResult, error) {
	res := sapi.InfoResult{}
	var err error
	res.Count, err = ig.ts.GetCount(rID, utils.RequestUser)
	return &res, err
}

func prepareInfoGetter(sp *mongodb.SessionProvider) (*infoGetter, error) {
	ts, err := mongodb.NewTextSaver(sp)
	if err != nil {
		return nil, errors.Wrap(err, "can't init text to DB saver")
	}
	return &infoGetter{ts: ts}, nil
}

func startPerfEndpoint() {
	port := goapp.Config.GetInt("debug.port")
	if port > 0 {
		goapp.Log.Info().Msgf("Starting Debug http endpoint at [::]:%d", port)
		portStr := strconv.Itoa(port)
		err := http.ListenAndServe(":"+portStr, nil)
		if err != nil {
			goapp.Log.Error().Err(errors.Wrap(err, "can't start Debug endpoint at "+portStr)).Send()
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

func initTracer(ctx context.Context, tracerURL string) (trace.TracerProvider, error) {
	if tracerURL == "" {
		log.Ctx(ctx).Warn().Msg("No tracer URL set, skipping OpenTelemetry initialization.")
		tp := noop.NewTracerProvider()
		otel.SetTracerProvider(tp)
		return tp, nil
	}

	propagator := propagation.NewCompositeTextMapPropagator(propagation.Baggage{}, propagation.TraceContext{})
	otel.SetTextMapPropagator(propagator)

	log.Ctx(ctx).Info().Str("url", tracerURL).Msg("Setting up OpenTelemetry")
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(tracerURL),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP HTTP exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewSchemaless(
			attribute.String("service.name", utils.ServiceName),
			attribute.String("service.version", version),
		)),
	)

	otel.SetTracerProvider(tp)
	return tp, nil
}

type shutdowner interface {
	Shutdown(ctx context.Context) error
}
